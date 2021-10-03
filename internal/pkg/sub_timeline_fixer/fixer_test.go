package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/james-bowman/nlp"
	"github.com/james-bowman/nlp/measures/pairwise"
	"gonum.org/v1/gonum/mat"
	"path"
	"strings"
	"testing"
)

func TestStopWordCounter(t *testing.T) {

	testDataPath := "../../../TestData/FixTimeline"
	testRootDir, err := pkg.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}

	subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())

	bFind, info, err := subParserHub.DetermineFileTypeFromFile(path.Join(testRootDir, "R&M S05E10 - English.srt"))
	if err != nil {
		t.Fatal(err)
	}
	if bFind == false {
		t.Fatal("not match sub types")
	}

	allString := strings.Join(info.OtherLines, " ")

	stopWords := StopWordCounter(strings.ToLower(allString), 5)

	print(len(stopWords))
	println(info.Name)
}

func TestGetOffsetTime(t *testing.T) {
	testDataPath := "../../../TestData/FixTimeline"
	testRootDir, err := pkg.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}

	enSubFile := path.Join(testRootDir, "R&M S05E01 - English.srt")
	ch_enSubFile := path.Join(testRootDir, "R&M S05E01 - 简英.srt")

	//enSubFile := path.Join(testRootDir, "R&M S05E10 - English.ass")
	//ch_enSubFile := path.Join(testRootDir, "R&M S05E10 - 简英.ass")

	time, err := GetOffsetTime(enSubFile, ch_enSubFile)
	if err != nil {
		return
	}

	print(time)
}

func TestTFIDF(t *testing.T) {
	testCorpus := []string{
		"The quick brown fox jumped over the lazy dog",
		"hey diddle diddle, the cat and the fiddle",
		"the cow jumped over the moon",
		"the little dog laughed to see such fun",
		"and the dish ran away with the spoon",
	}

	query := "the brown fox ran around the dog"

	vectoriser := nlp.NewCountVectoriser(StopWords...)
	transformer := nlp.NewTfidfTransformer()

	// set k (the number of dimensions following truncation) to 4
	reducer := nlp.NewTruncatedSVD(4)

	lsiPipeline := nlp.NewPipeline(vectoriser, transformer, reducer)

	// Transform the corpus into an LSI fitting the model to the documents in the process
	lsi, err := lsiPipeline.FitTransform(testCorpus...)
	if err != nil {
		fmt.Printf("Failed to process documents because %v", err)
		return
	}

	// run the query through the same pipeline that was fitted to the corpus and
	// to project it into the same dimensional space
	queryVector, err := lsiPipeline.Transform(query)
	if err != nil {
		fmt.Printf("Failed to process documents because %v", err)
		return
	}

	// iterate over document feature vectors (columns) in the LSI matrix and compare
	// with the query vector for similarity.  Similarity is determined by the difference
	// between the angles of the vectors known as the cosine similarity
	highestSimilarity := -1.0
	var matched int
	_, docs := lsi.Dims()
	for i := 0; i < docs; i++ {
		similarity := pairwise.CosineSimilarity(queryVector.(mat.ColViewer).ColView(0), lsi.(mat.ColViewer).ColView(i))
		if similarity > highestSimilarity {
			matched = i
			highestSimilarity = similarity
		}
	}

	fmt.Printf("Matched '%s'", testCorpus[matched])
	// Output: Matched 'The quick brown fox jumped over the lazy dog'
}
