package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

//go:embed tpl
var tpl embed.FS

//go:embed static
var static embed.FS
var (
	notePath                string
	homeTpl, viewTpl, mdTpl *template.Template
	prefixLen               int
	homeText                = &atomic.Value{}
)

func init() {
	exPath, _ := os.Executable()
	var err error
	if strings.Contains(exPath, "/go-build") || strings.Contains(exPath, "/___go_build_") {
		exPath, err = os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("failed to get working directory: %v", err))
		}
	}

	parseTemplate := func(name string) (t *template.Template, err error) {
		t, err = template.ParseFS(tpl, "tpl/"+name)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		return t, nil
	}

	if homeTpl, err = parseTemplate("home.html"); err != nil {
		panic(err)
	}

	if viewTpl, err = parseTemplate("view.html"); err != nil {
		panic(err)
	}

	if mdTpl, err = parseTemplate("md.html"); err != nil {
		panic(err)
	}

	notePath = exPath[:strings.LastIndex(exPath, "/server")] + "/notefile"
	prefixLen = len(notePath) + 1
	fmt.Println("notePath:", notePath, "prefixLen:", prefixLen)

	load()
	go func() {
		t := time.Tick(time.Minute)
		for range t {
			load()
		}
	}()
}

func main() {
	http.Handle("/static/", http.FileServer(http.FS(static)))
	//home
	http.HandleFunc("/{$}", home)
	http.HandleFunc("/view/{path}", view)
	server := &http.Server{Addr: ":1024"}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal("ListenAndServe.err:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatal("Server Shutdown.err:", err)
	}
}

func load() {
	buf := &strings.Builder{}
	read(notePath, buf)
	homeText.Store(template.HTML(buf.String()))
}

func read(path string, w io.Writer) {
	dirs, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("read.err:", err)
		return
	}
	for _, d := range dirs {
		w.Write([]byte("<li>"))
		p := path + "/" + d.Name()
		if d.IsDir() {
			w.Write([]byte(`<span class="dir"><span>📘</span> `))
			w.Write([]byte(d.Name()))
			w.Write([]byte(`</span><ul class="sub-ul">`))
			read(p, w)
			w.Write([]byte("</ul>"))
		} else {
			w.Write([]byte(`<a class="file" href="/view/`))
			w.Write([]byte(url.PathEscape(p[prefixLen:])))
			w.Write([]byte(`"><small>📄</small> `))
			w.Write([]byte(d.Name()))
			w.Write([]byte("</a>"))
		}
		w.Write([]byte("</li>\n"))
	}
}

func home(w http.ResponseWriter, _ *http.Request) {
	homeTpl.Execute(w, homeText.Load())
}

type ViewData struct {
	Title      string
	Nav        string
	Content    string
	IsMarkdown bool
}

func view(w http.ResponseWriter, r *http.Request) {
	p := r.PathValue("path")
	if len(p) < 2 {
		http.NotFound(w, r)
		return
	}
	b, err := os.ReadFile(notePath + "/" + p)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	isMarkdown := strings.HasSuffix(p, ".md")
	//fmt.Println("isMarkdown:", isMarkdown, p)
	if isMarkdown {
		// 使用 Markdown 模板处理 .md 文件
		mdTpl.Execute(w, &ViewData{
			Title:   p,
			Nav:     strings.ReplaceAll(p, "/", "📌"),
			Content: string(b),
		})
	} else {
		// 使用普通视图模板处理其他文件
		viewTpl.Execute(w, &ViewData{
			Title:   p,
			Nav:     strings.ReplaceAll(p, "/", "📌"),
			Content: string(b),
		})
	}
}
