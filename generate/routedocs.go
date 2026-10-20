package generate

import (
	"path/filepath"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/ast"
	"github.com/zeromicro/go-zero/tools/goctl/pkg/parser/api/parser"
)

// routeMethodDocs recovers the comments written directly above a route's
// method, keyed by handler name, from apiFile and the files it imports:
//
//	@handler getUserInfo
//	/* @respdoc-500 (ErrorResponse) // Server Error */
//	get /user/:id (UserInfoReq) returns (UserInfoReply)
//
// goctl's parser attaches these to the method token, but its analyzer only
// copies the comments above @doc/@handler into spec.Route, so they never reach
// the plugin. A file that fails to parse is skipped: parser.Parse has already
// accepted it, and the comments are best-effort.
func routeMethodDocs(apiFile string) map[string][]string {
	docs := map[string][]string{}
	if apiFile == "" {
		return docs
	}
	seen := map[string]bool{}
	var walk func(filename string)
	walk = func(filename string) {
		if abs, err := filepath.Abs(filename); err == nil {
			filename = abs
		}
		if seen[filename] {
			return
		}
		seen[filename] = true

		p := parser.New(filename, "")
		tree := p.Parse()
		if p.CheckErrors() != nil || tree == nil {
			return
		}
		var imports []string
		for _, stmt := range tree.Stmts {
			switch s := stmt.(type) {
			case *ast.ImportLiteralStmt:
				imports = append(imports, s.Value.Token.Text)
			case *ast.ImportGroupStmt:
				for _, v := range s.Values {
					imports = append(imports, v.Token.Text)
				}
			case *ast.ServiceStmt:
				for _, item := range s.Routes {
					if item.AtHandler == nil || item.Route == nil || item.Route.Method == nil {
						continue
					}
					if list := item.Route.Method.HeadCommentGroup.List(); len(list) > 0 {
						name := item.AtHandler.Name.Token.Text
						docs[name] = append(docs[name], list...)
					}
				}
			}
		}
		// import paths resolve relative to the importing file, as in goctl
		for _, imp := range imports {
			imp = strings.ReplaceAll(imp, `"`, "")
			if !filepath.IsAbs(imp) {
				imp = filepath.Join(filepath.Dir(filename), imp)
			}
			walk(imp)
		}
	}
	walk(apiFile)
	return docs
}

// routeDocs returns every comment attached to a route that may carry
// annotations such as @respdoc-, without duplicates.
func routeDocs(route spec.Route, methodDocs map[string][]string) []string {
	var out []string
	seen := map[string]bool{}
	for _, list := range [][]string{route.Doc, route.HandlerDoc, methodDocs[route.Handler], route.HandlerComment, route.Comment} {
		for _, doc := range list {
			if !seen[doc] {
				seen[doc] = true
				out = append(out, doc)
			}
		}
	}
	return out
}
