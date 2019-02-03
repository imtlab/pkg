package urlquerybuilder

import (
	"strings"
)

/*
	Package urlquerybuilder provides a handy way to build up URL Query strings
*/

type UrlQueryProperty struct{
	/*	Note: Key and Value must be exported because to supply a UrlQueryProperty struct literal
		to UrlQueryBuilder.Append() does implicit assignments to these fields.
	*/
	Key		string
	Value	string
}

type UrlQueryBuilder []UrlQueryProperty


func (p *UrlQueryProperty) String() string {
	return p.Key + "=" + p.Value
}

/*	because Append() modifies the header (i.e., the length) of the calling slice,
	the receiver must be a pointer or our change won't propagate back to the caller.
*/
func (p *UrlQueryBuilder) Append(moreProps ...UrlQueryProperty) {
	*p = append(*p, moreProps...)
}

func (xProps UrlQueryBuilder) String() string {
	//	Too bad I have to create an interim slice since Go has no no linq-like closure to execute for each entry: prop => prop.String()
	xs := make([]string, 0, len(xProps))
	for _, prop := range xProps {
		xs = append(xs, prop.String())
	}
	return strings.Join(xs, "&")
}
