package builtin

import (
	"context"
	"os"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
)

const (
	templateFileTag = rocket.NamedIO("template")
)

type (
	// Template is a task to expand the input file using
	// the Template file. Output is written to the redirected STDOUT
	// Delims is used to change the standard golang templatiing delimiters
	// This can be useful when processing a source file that itself uses golang tempalting.
	Template struct {
		// Template file
		FileTemplate      string `mapstructure:"template"`
		InlineTemplate    string `mapstructure:"inline"`
		rocket.OutputSpec `mapstructure:",squash"`
		Delims            rocket.Delims `mapstructure:"delims"`
	}

	templateType struct{}
)

func (templateType) Type() string {
	return "template"
}

func validateTemplateConfig(ctx context.Context, capComm *rocket.CapComm, templateCfg *Template) error {
	if templateCfg.FileTemplate != "" && templateCfg.InlineTemplate != "" {
		return errors.New("both a file and inline template have been specified, only one is allowed")
	} else if templateCfg.FileTemplate == "" && templateCfg.InlineTemplate == "" {
		return errors.New("neither a file or inline template have been specified, please provide one of them")
	}

	// Expand the template file name
	if templateCfg.FileTemplate != "" {
		if err := capComm.AddFile(ctx, templateFileTag, templateCfg.FileTemplate, rocket.IOModeInput); err != nil {
			return errors.Wrap(err, "expanding template file name")
		}
	}

	// Expand redirect settings into cap Comm
	if err := capComm.AttachOutput(ctx, templateCfg.OutputSpec); err != nil {
		return errors.Wrap(err, "expanding output settings")
	}

	return nil
}

func (templateType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {
	templateCfg := &Template{}

	if err := mapstructure.Decode(task.Definition, templateCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	if err := validateTemplateConfig(ctx, capComm, templateCfg); err != nil {
		return nil, err
	}

	fn := func(runCtx context.Context) error {
		//	Load the template
		t, err := loadTemplate(capComm, task.Name, templateCfg)
		if err != nil {
			return errors.Wrap(err, "create template")
		}

		writer, cf, err := setupOutput(capComm)
		defer cf.Close()
		if err != nil {
			return err
		}

		if err := t.Execute(writer, capComm.GetTemplateData(ctx)); err != nil {
			return errors.Wrap(err, "execute template")
		}

		return nil
	}

	return fn, nil
}

func loadTemplate(capComm *rocket.CapComm, name string, templateCfg *Template) (*template.Template, error) {
	var tt string
	if templateCfg.InlineTemplate != "" {
		tt = templateCfg.InlineTemplate
	} else if b, err := capComm.GetFileDetails(templateFileTag).ReadFile(); err != nil {
		return nil, errors.Wrap(err, "read template file")
	} else {
		tt = string(b)
	}

	d := templateCfg.Delims

	t, err := template.New(name).
		Funcs(capComm.FuncMap()).
		Delims(d.Left, d.Right).
		Parse(tt)
	if err != nil {
		return nil, errors.Wrap(err, "parse template")
	}

	return t, err
}

func setupOutput(capComm *rocket.CapComm) (*os.File, closeFiles, error) {
	cf := make(closeFiles, 0, 3)

	// Handle output
	outFd := capComm.GetFileDetails(rocket.OutputIO)
	if outFd != nil {
		outFile, err := outFd.OpenOutput()
		if err != nil {
			return nil, cf, errors.Wrap(err, string(rocket.OutputIO))
		}
		cf = append(cf, outFile)
		return outFile, cf, nil
	}

	return os.Stdout, cf, nil
}

func init() {
	rocket.Default().RegisterTaskTypes(templateType{})
}
