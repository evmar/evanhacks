package transcode

import (
	"os"
	"fmt"
	"testing"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func TestRender(t *testing.T) {
	f, err := os.Open("andy.json")
	check(err)
	feed, err := ReadGPFeed(f)
	check(err)
	check(f.Close())

	f, err = os.Create("andy.html")
	check(err)
	fmt.Fprintf(f, `<!doctype html><meta charset=utf-8>
<head>
<link rel=stylesheet href=style.css>
</head>
`)
	for _, item := range feed.Items {
		fmt.Fprintf(f, "<div class=post>%s</div>\n", RenderPost(item))
	}
	check(f.Close())
}
