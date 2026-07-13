package utils

import "testing"

func TestId(t *testing.T) {
	a := "abc"
	b := mangle([]byte(a))
	t.Log(b, string(b))

	c := mangle(b)
	t.Log(c, string(c))

	d := simpleCipher([]byte(a))
	t.Log(d, string(d))
	e := simpleCipher(d)
	t.Log(e, string(e))
}

func TestReplace(t *testing.T) {
	initReplaycers()

	t.Log(toAbbr("sr:match:123456"))

	t.Log(fromAbbr("{1}"))

	aa := Encode("sr:match:123456")
	t.Log(aa)
	t.Log(Decode(aa))
}
