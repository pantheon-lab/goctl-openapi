package action

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/zeromicro/go-zero/tools/goctl/api/parser"
	"github.com/zeromicro/go-zero/tools/goctl/plugin"
	"github.com/pantheon-lab/goctl-openapi/generate"
)

func Generator(ctx *cli.Context) error {
	apiFile := ctx.String("api")
	basepath := ctx.String("basepath")
	host := ctx.String("host")
	schemes := ctx.String("schemes")

	if apiFile != "" {
		return standaloneGenerate(apiFile, host, basepath, schemes, ctx)
	}
	return pluginGenerate(host, basepath, schemes, ctx)
}

func standaloneGenerate(apiFile, host, basepath, schemes string, ctx *cli.Context) error {
	data, err := os.ReadFile(apiFile)
	if err != nil {
		return err
	}

	api, err := parser.ParseContent(string(data), apiFile)
	if err != nil {
		return err
	}

	dir := ctx.String("dir")
	if dir == "" {
		dir = filepath.Dir(apiFile)
	}

	p := &plugin.Plugin{
		Api:         api,
		ApiFilePath: apiFile,
		Dir:         dir,
	}

	fileName := ctx.String("filename")
	if fileName == "" {
		in := filepath.Base(apiFile)
		if ext := filepath.Ext(in); ext != "" {
			in = strings.TrimSuffix(in, ext)
		}
		fileName = in + ".yaml"
	}

	return generate.Do(fileName, host, basepath, schemes, p)
}

func pluginGenerate(host, basepath, schemes string, ctx *cli.Context) error {
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	var info struct {
		ApiFilePath string
		Style       string
		Dir         string
	}
	if err := json.Unmarshal(content, &info); err != nil {
		return err
	}

	p := &plugin.Plugin{
		ApiFilePath: info.ApiFilePath,
		Style:       info.Style,
		Dir:         info.Dir,
	}

	if info.ApiFilePath != "" {
		data, err := os.ReadFile(info.ApiFilePath)
		if err != nil {
			return err
		}
		api, err := parser.ParseContent(string(data), info.ApiFilePath)
		if err != nil {
			return err
		}
		p.Api = api
	}

	fileName := ctx.String("filename")
	if fileName == "" {
		fileName = "rest.openapi.yaml"
		if p.ApiFilePath != "" {
			in := filepath.Base(p.ApiFilePath)
			if ext := filepath.Ext(in); ext != "" {
				in = strings.TrimSuffix(in, ext)
			}
			fileName = in + ".yaml"
		}
	}

	return generate.Do(fileName, host, basepath, schemes, p)
}
