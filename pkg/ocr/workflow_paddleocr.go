package ocr

import (
	"log"

	classifypaddleocr "github.com/multippt/gopaddleocr/pkg/ocr/classify/paddleocr"
	"github.com/multippt/gopaddleocr/pkg/ocr/common"
	"github.com/multippt/gopaddleocr/pkg/ocr/detect"
	"github.com/multippt/gopaddleocr/pkg/ocr/detect/boxmerge"
	detectpaddleocr "github.com/multippt/gopaddleocr/pkg/ocr/detect/paddleocr"
	ppdoclayoutv3 "github.com/multippt/gopaddleocr/pkg/ocr/detect/pp-doclayoutv3"
	recognizepaddleocr "github.com/multippt/gopaddleocr/pkg/ocr/recognize/paddleocr"
	ort "github.com/yalue/onnxruntime_go"
)

func NewPaddleOCRWorkflow(engineConf *Config, configSrc common.ConfigSource) *Workflow {
	sessionOptions, err := ort.NewSessionOptions()
	if err != nil {
		panic(err)
	}
	_ = sessionOptions.SetIntraOpNumThreads(1)

	var det detect.Detector
	if engineConf.EnableBoxMerge {
		log.Println("EnableBoxMerge")
		det = boxmerge.NewModel(
			ppdoclayoutv3.NewModel().SetSessionOptions(sessionOptions),
			detectpaddleocr.NewModel().SetSessionOptions(sessionOptions),
		)
	}

	return NewWorkflow(
		det,
		classifypaddleocr.NewModel().SetSessionOptions(sessionOptions),
		recognizepaddleocr.NewModel().SetSessionOptions(sessionOptions),
		configSrc,
	)
}
