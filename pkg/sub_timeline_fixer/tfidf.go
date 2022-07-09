package sub_timeline_fixer

import (
	"errors"
	"fmt"

	"github.com/james-bowman/nlp"
	"gonum.org/v1/gonum/mat"
)

// NewTFIDF 初始化 TF-IDF
func NewTFIDF(testCorpus []string) (*nlp.Pipeline, mat.Matrix, error) {
	vectors := nlp.NewCountVectoriser(EnStopWords...)
	transformer := nlp.NewTfidfTransformer()
	// set k (the number of dimensions following truncation) to 4
	reducer := nlp.NewTruncatedSVD(4)
	lsiPipeline := nlp.NewPipeline(vectors, transformer, reducer)
	// Transform the corpus into an LSI fitting the model to the documents in the process
	lsi, err := lsiPipeline.FitTransform(testCorpus...)
	if err != nil {
		return nil, lsi, errors.New(fmt.Sprintf("Failed to process testCorpus documents because %v", err))
	}

	return lsiPipeline, lsi, nil
}
