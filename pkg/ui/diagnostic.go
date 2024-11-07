package mainui

import "github.com/tectiv3/go-lsp"

type editor_diagnostic struct {
	data lsp.PublishDiagnosticsParams
}

func NewLspDiagnostic(diags lsp.PublishDiagnosticsParams) *editor_diagnostic {
	return &editor_diagnostic{
		data: diags,
	}
}

type project_diagnostic struct {
	data []editor_diagnostic
}

func (prj *project_diagnostic) Update(diags lsp.PublishDiagnosticsParams) {
	for i, v := range prj.data {
		if v.data.URI.AsPath().String() == diags.URI.AsPath().String() {
			if diags.IsClear {
				data := prj.data[:i]
				data = append(data, prj.data[i+1:]...)
				prj.data = data
				return
			} else {
				prj.data[i].data = diags
				return
			}
		}
	}
	if !diags.IsClear {
		prj.data = append(prj.data, *NewLspDiagnostic(diags))
	}
}
