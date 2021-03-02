package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"text/template"

	"github.com/micro/micro/v3/service/errors"
	"github.com/nikitaksv/gendata/handler/validation"
	pb "github.com/nikitaksv/gendata/proto"

	log "github.com/micro/micro/v3/service/logger"
	"github.com/nikitaksv/gendata/handler/meta"
)

const ErrValidation = "ERROR_VALIDATION"

type gendata struct {
	logger log.Logger
}

func New(logger log.Logger) pb.GendataHandler {
	return &gendata{logger: logger}
}

// Call is a single request handler called via client.Call or the generated client code
func (g *gendata) Generate(ctx context.Context, req *pb.GenerateRequest, rsp *pb.GenerateResponse) error {
	err := validation.ValidateGenerateRequest(req)
	if err != nil {
		errorBs, _ := json.Marshal(err)
		return errors.BadRequest(ErrValidation, "%s", string(errorBs))
	}

	// Create typeAliases
	tA := typeAliases(req.Schema.Options.TypeAliases)

	// Create Meta
	m := meta.Meta{
		SortProps:       req.Schema.Options.SortSchemaFields,
		Key:             meta.Key(req.Schema.Options.ClassPrefix + req.Schema.Options.RootName),
		PrefixObjectKey: req.Schema.Options.ClassPrefix,
		Type:            meta.NewType(meta.TypeString, tA.Apply(meta.TypeString)),
		TypeAliases:     tA,
		Properties:      nil,
	}
	if m.Key.String() == "" {
		m.Key = "RootClass"
	}

	// Parse Content
	tmpl, err := template.New(req.Schema.Options.RootName).Parse(req.Template.Content)
	if err != nil {
		return err
	}

	// Parse Content
	err = json.Unmarshal([]byte(req.Schema.Content), &m)
	if err != nil {
		return err
	}

	var metas []meta.Meta
	if req.Schema.Options.SplitObjectsIntoFiles {
		metas = meta.Split(m)
	} else {
		metas = append(metas, m)
	}

	files := make([]*pb.File, len(metas))
	buf := &bytes.Buffer{}
	for i, el := range metas {
		// Generate Content
		err = tmpl.Execute(buf, el)
		if err != nil {
			return err
		}
		files[i] = &pb.File{
			Name:      el.Key.PascalCase().String(),
			Extension: req.Schema.Type,
			Content:   buf.String(),
		}
		buf.Reset()
	}

	rsp.Files = files

	return nil
}

func typeAliases(tA *pb.Types) meta.TypeAliases {
	return meta.TypeAliases{
		"null":        tA.Null,
		"int":         tA.Int,
		"string":      tA.String_,
		"bool":        tA.Bool,
		"float":       tA.Float,
		"object":      tA.Object,
		"array":       tA.Array,
		"arrayObject": tA.ArrayObject,
		"arrayInt":    tA.ArrayInt,
		"arrayString": tA.ArrayString,
		"arrayBool":   tA.ArrayBool,
		"arrayFloat":  tA.ArrayFloat,
	}
}
