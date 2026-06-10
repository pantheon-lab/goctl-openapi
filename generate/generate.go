package generate

import (
	"bytes"
	"os"

	"github.com/zeromicro/go-zero/tools/goctl/plugin"
	"gopkg.in/yaml.v3"
)

func Do(filename string, host string, basePath string, schemes string, in *plugin.Plugin) error {
	o, err := applyGenerate(in, host, basePath, schemes)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(o); err != nil {
		return err
	}
	if err := encoder.Close(); err != nil {
		return err
	}

	output := in.Dir + "/" + filename
	return os.WriteFile(output, buf.Bytes(), 0666)
}
