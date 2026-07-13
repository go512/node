package Tc

type MatchesRep[T any] struct {
	Code    int        `json:"code"`
	Results []Match[T] `json:"results"`
}

type Match[T any] struct {
	MatchId   string `json:"match_id"`
	MatchName string `json:"match_name"`
}

type TsDoc interface {
	GetId() string
	GetUpdatedAt() int32
	// ToPb() proto.Message
	GetTsId() string
	GetSportId() int32
	ToBytes() []byte
}

// TsDocMinIn 可以通过内嵌的TsDocMinIn方式实现TsDoc接口
type TsDocMinIn struct {
	Id        string `json:"id"`
	UpdatedAt int32  `json:"updated_at"`
}

func (d *TsDocMinIn) GetUpdatedAt() int32 {
	return d.UpdatedAt
}

func (d *TsDocMinIn) GetSportId() int32 {
	// Todo
	return 1
}

// TsDocMatchMixin 可以通过内嵌的TsDocMinIn方式实现TsDoc接口 所以GetSportId，GetUpdatedAt方法可以直接使用
type TsDocMatchMixin struct {
	TsDocMinIn
	MatchId   string `json:"match_id"`
	MatchName string `json:"match_name"`
}

func (d *TsDocMatchMixin) GetId() string {
	return d.MatchId
}

type TsMatchSoccerBasketball struct {
	TsDocMatchMixin
	HomeTeam string `json:"home_team"`
	AwayTeam string `json:"away_team"`
}

func (d *TsMatchSoccerBasketball) ToBytes() []byte {
	panic("not implement")
}

type TsDocSeasonMixin struct {
	TsDocMinIn
	SeasonId   string `json:"season_id"`
	SeasonName string `json:"season_name"`
}

type TsSeasonSoccerBasketball struct {
	TsDocMinIn
	CompetitionId string `json:"competition_id"`
	Year          string `json:"year"`
}

func (d *TsSeasonSoccerBasketball) ToBytes() []byte {
	// TODO implement me
	panic("implement me")
}

func (d *TsSeasonSoccerBasketball) GetTsId() string {
	return d.Id
}

type TsDocPlayerMixin struct {
	TsDocMinIn
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type TsPlayer struct {
	TsDocPlayerMixin
	PlayerId string `json:"player_id"`
	Age      int32  `json:"age"`
	Weight   int32  `json:"weight"`
	Height   int32  `json:"height"`
}

func (d *TsPlayer) ToBytes() []byte {
	// TODO implement me
	panic("implement me")
}
