package server

import (
	"issue-man/config"
	"issue-man/global"
	"testing"
)

func Test_genBody(t *testing.T) {
	type args struct {
		remove  bool
		file    string
		oldBody string
	}
	tests := []struct {
		name       string
		args       args
		wantBody   *string
		wantLength int
	}{
		{
			name: "gen-1",
			args: args{
				remove:  false,
				file:    "content/en/faq/_index.html",
				oldBody: "\"## Source\n\n#### Files\n\n- https://github.com/istio/istio.io/tree/master/content/en/_index.html\n## Translate\n\n#### Files\n\n- https://github.com/1kib/new/tree/master/content/zh/_index.html\n\"",
			},
			wantBody:   nil,
			wantLength: 0,
		},
	}
	global.Conf = &config.Config{}
	global.Conf.Repository.Spec.Source.RemovePrefix = "content/en"
	global.Conf.Repository.Spec.Source.Site = "istio.io/latest"
	global.Conf.Repository.Spec.Translate.Site = "istio.io/latest/zh"
	global.Conf.Repository.Spec.Source.Owner = "istio"
	global.Conf.Repository.Spec.Source.Repository = "istio.io"
	global.Conf.Repository.Spec.Translate.Owner = "istio"
	global.Conf.Repository.Spec.Translate.Repository = "istio.io"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBody, gotLength := genBody(tt.args.remove, tt.args.file, tt.args.oldBody)
			t.Logf("body: %s\n\nlength:%d\n", *gotBody, gotLength)
			//if !reflect.DeepEqual(gotBody, tt.wantBody) {
			//	t.Errorf("genBody() gotBody = %v, want %v", gotBody, tt.wantBody)
			//}
			//if gotLength != tt.wantLength {
			//	t.Errorf("genBody() gotLength = %v, want %v", gotLength, tt.wantLength)
			//}
		})
	}
}
