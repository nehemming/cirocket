package builtin

import (
	"context"
	"io"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/nehemming/cirocket/pkg/providers"
	"github.com/nehemming/cirocket/pkg/rocket"
	"github.com/pkg/errors"
)

const (
	templateResourceID = providers.ResourceID("template")
)

type (
	// Template is a task to expand the input file using
	// the Template file. Output is written to the redirected STDOUT
	// Delims is used to change the standard golang templatiing delimiters
	// This can be useful when processing a source file that itself uses golang tempalting.
	Template struct {
		Template rocket.InputSpec `mapstructure:"template"`

		// OutputSpec is the specification for the template output
		Output *rocket.OutputSpec `mapstructure:"output"`

		// Delims are the delimiters used to identify template script.
		// Leave blank for the default go templating delimiters
		Delims rocket.Delims `mapstructure:"delims"`
	}

	templateType struct{}
)

func (templateType) Type() string {
	return "template"
}

func configureSources(ctx context.Context, capComm *rocket.CapComm, templateCfg *Template) error {
	// Preevent inline being expanded as input to a template
	if templateCfg.Template.Inline != "" {
		templateCfg.Template.SkipExpand = true
	}
	if err := capComm.AttachInputSpec(ctx, templateResourceID, templateCfg.Template); err != nil {
		return errors.Wrap(err, "template")
	}

	if templateCfg.Output != nil {
		if err := capComm.AttachOutputSpec(ctx, rocket.OutputIO, *templateCfg.Output); err != nil {
			return errors.Wrap(err, "output")
		}
	}
	return nil
}

func (templateType) Prepare(ctx context.Context, capComm *rocket.CapComm, task rocket.Task) (rocket.ExecuteFunc, error) {
	templateCfg := &Template{}

	if err := mapstructure.Decode(task.Definition, templateCfg); err != nil {
		return nil, errors.Wrap(err, "parsing template type")
	}

	fn := func(runCtx context.Context) error {
		// Late configure sources, allow previous steps to be available
		if err := configureSources(runCtx, capComm, templateCfg); err != nil {
			return err
		}

		//	Load the template
		t, err := loadTemplate(runCtx, capComm, task.Name, templateCfg)
		if err != nil {
			return errors.Wrap(err, "template")
		}

		// Prepare output
		outputResource := capComm.GetResource(rocket.OutputIO)
		writer, err := outputResource.OpenWrite(runCtx)
		if err != nil {
			return errors.Wrap(err, "output")
		}
		defer writer.Close()

		// Execute
		if err := t.Execute(writer, capComm.GetTemplateData(ctx)); err != nil {
			return errors.Wrap(err, "template")
		}

		return nil
	}

	return fn, nil
}

func loadTemplate(ctx context.Context, capComm *rocket.CapComm, name string, templateCfg *Template) (*template.Template, error) {
	// Get template data
	templateResource := capComm.GetResource(templateResourceID)
	r, err := templateResource.OpenRead(ctx)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return template.New(name).
		Funcs(capComm.FuncMap()).
		Delims(templateCfg.Delims.Left, templateCfg.Delims.Right).Parse(string(b))
}

func init() {
	rocket.Default().RegisterTaskTypes(templateType{})
}
