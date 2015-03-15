package main

import (
	"fmt"
	"github.com/bronze1man/kmg/kmgConsole"
	"github.com/bronze1man/kmg/kmgFile"
	"github.com/bronze1man/kmg/kmgHtmlTemplate"
	"html/template"
	"kmgBlog/internal/MarkDown"
	"net/http"
	"path/filepath"
	"strings"
)

type article struct {
	Path    string
	Title   string
	Content string
}

type kmgBlog struct {
	articleList []article
}

const articleListTpl = `<html><head><meta charset="utf-8">
        <title>bronze1man's blog</title>
	<style>
table, th, td {
    border: 1px solid black;
}
</style>
</head><body>
    <h1>bronze1man's blog</h1>
    <form action="" method="GET" >
        <input type="hidden" name="n" value="SearchAction"/>
        <input type="text" placeholder="search" name="Keyword"/>
        <input type="submit"/>
    </form>
    <hr/>
    <table>
		<tbody>
			{{range .}}
			<tr>
			    <td><a href="/?n=ArticlePage&Path={{.Path}}">{{.Path}}</a></td>
			</tr>
			{{end}}
		</tbody>
	</table>
	</body>
    </html>`

func (s *kmgBlog) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	n := req.URL.Query().Get("n")
	if n == "" {
		n = "IndexPage"
	}
	switch n {
	case "IndexPage":
		kmgHtmlTemplate.MustRenderToWriter(w, articleListTpl, s.articleList)
		return
	case "ArticlePage":
		path := req.URL.Query().Get("Path")
		var art article
		for i := range s.articleList {
			if s.articleList[i].Path == path {
				art = s.articleList[i]
			}
		}
		if art.Path == "" {
			http.Redirect(w, req, "/", 301)
			return
		}
		output := MarkDown.MarkdownCommon([]byte(art.Content))
		kmgHtmlTemplate.MustRenderToWriter(w, `<html><head><meta charset="utf-8">
        <title>{{ .Article.Title }}</title>
<style>
table, th, td,pre {
    border: 1px solid black;
}
</style>
</head>
    <h1>{{ .Article.Title }}</h1>
    <hr>
    <div>{{.Html}}</div>
    <hr>
    <a href="#" onclick="document.getElementById('Origin').setAttribute('style','');">Show Origin Markdown</a>
    <pre id="Origin" style="display:none;">{{.Article.Content}}</pre>
    </html>`, struct {
			Article article
			Html    template.HTML
		}{
			Article: art,
			Html:    template.HTML(output),
		})
		return
	case "SearchAction":
		keyword := req.URL.Query().Get("Keyword")
		if keyword == "" {
			http.Redirect(w, req, "/", 301)
			return
		}
		outArticleList := []article{}
		for _, thisArt := range s.articleList {
			if strings.Contains(thisArt.Path, keyword) {
				outArticleList = append(outArticleList, thisArt)
				continue
			}
			if strings.Contains(thisArt.Content, keyword) {
				outArticleList = append(outArticleList, thisArt)
				continue
			}
		}
		kmgHtmlTemplate.MustRenderToWriter(w, articleListTpl, outArticleList)
		return
	default:
		http.NotFound(w, req)
		return
	}
}

func main() {
	//加载所有markdown文件,并且转成成html放到内存中
	rootPath := "/Users/bronze1man/material/it-summary"
	files, err := kmgFile.GetAllFiles(rootPath)
	kmgConsole.ExitOnErr(err)
	articleList := []article{}
	for _, path := range files {
		if filepath.Ext(path) != ".md" {
			continue
		}
		Content := kmgFile.MustReadFileAll(path)
		Path, err := filepath.Rel(rootPath, path)
		if err != nil {
			panic(err)
		}
		Title := kmgFile.PathBaseWithoutExt(path)
		articleList = append(articleList, article{
			Path:    kmgFile.PathTrimExt(Path),
			Title:   Title,
			Content: string(Content),
		})
	}
	fmt.Println("read", len(articleList), "files")
	s := &kmgBlog{
		articleList: articleList,
	}
	err = http.ListenAndServe(":20006", s)
	kmgConsole.ExitOnErr(err)
	//使用渲染
	//添加搜索功能
}
