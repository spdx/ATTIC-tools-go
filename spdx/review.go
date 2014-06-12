package spdx

type Review struct {
	Reviewer ValueStr // mandatory
	Date     ValueStr // mandatory
	Comment  ValueStr // optional
}

func (a *Review) Equal(b *Review) bool {
	return a.Reviewer.Val == b.Reviewer.Val && a.Date.Val == b.Date.Val && a.Comment.Val == b.Comment.Val
}
