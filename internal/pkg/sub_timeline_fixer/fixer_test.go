package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
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

	subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())

	type args struct {
		enSubFile              string
		ch_enSubFile           string
		staticLineFileSavePath string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{name: "R&M S05E01", args: args{enSubFile: path.Join(testRootDir, "R&M S05E01 - English.srt"),
			ch_enSubFile:           path.Join(testRootDir, "R&M S05E01 - 简英.srt"),
			staticLineFileSavePath: "bar.html"}, want: -6.42981818181818, wantErr: false},
		{name: "R&M S05E10", args: args{enSubFile: path.Join(testRootDir, "R&M S05E10 - English.ass"),
			ch_enSubFile:           path.Join(testRootDir, "R&M S05E10 - 简英.ass"),
			staticLineFileSavePath: "bar.html"}, want: -6.335985401459854, wantErr: false},
		{name: "R&M S05E10-shooter", args: args{enSubFile: path.Join(testRootDir, "R&M S05E10 - English.ass"),
			ch_enSubFile:           path.Join(testRootDir, "R&M S05E10 - 简英-shooter.ass"),
			staticLineFileSavePath: "bar.html"}, want: -6.335985401459854, wantErr: false},
		{name: "基地 S01E03", args: args{enSubFile: path.Join(testRootDir, "基地 S01E03 - English.ass"),
			ch_enSubFile:           path.Join(testRootDir, "基地 S01E03 - 简英.ass"),
			staticLineFileSavePath: "bar.html"}, want: -32.09061538461539, wantErr: false},

		{name: "Dan Brown's The Lost Symbol - S01E01", args: args{
			enSubFile:              path.Join(testRootDir, tmpSubDataFolderName, "Dan Brown's The Lost Symbol - S01E01 - As Above, So Below WEBDL-720p", "Dan Brown's The Lost Symbol - S01E01 - As Above, So Below WEBDL-720p.chinese(inside).ass"),
			ch_enSubFile:           path.Join(testRootDir, tmpSubDataFolderName, "Dan Brown's The Lost Symbol - S01E01 - As Above, So Below WEBDL-720p", "Dan Brown's The Lost Symbol - S01E01 - As Above, So Below WEBDL-720p.chinese(简英,shooter).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 1.3217821782178225, wantErr: false},
		{name: "Dan Brown's The Lost Symbol - S01E02", args: args{
			enSubFile:              path.Join(testRootDir, tmpSubDataFolderName, "Dan Brown's The Lost Symbol - S01E02 - The Araf WEBDL-1080p", "Dan Brown's The Lost Symbol - S01E02 - The Araf WEBDL-1080p.chinese(inside).ass"),
			ch_enSubFile:           path.Join(testRootDir, tmpSubDataFolderName, "Dan Brown's The Lost Symbol - S01E02 - The Araf WEBDL-1080p", "Dan Brown's The Lost Symbol - S01E02 - The Araf WEBDL-1080p.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: -0.5253383458646617, wantErr: false},
		{name: "Dan Brown's The Lost Symbol - S01E03", args: args{
			enSubFile:              path.Join(testRootDir, tmpSubDataFolderName, "Dan Brown's The Lost Symbol - S01E03 - Murmuration WEBDL-1080p", "Dan Brown's The Lost Symbol - S01E03 - Murmuration WEBDL-1080p.chinese(inside).ass"),
			ch_enSubFile:           path.Join(testRootDir, tmpSubDataFolderName, "Dan Brown's The Lost Symbol - S01E03 - Murmuration WEBDL-1080p", "Dan Brown's The Lost Symbol - S01E03 - Murmuration WEBDL-1080p.chinese(简英,shooter).ass"),
			staticLineFileSavePath: "bar.html"},
			want: -0.505656, wantErr: false},
		{name: "Dan Brown's The Lost Symbol - S01E03", args: args{
			enSubFile:              path.Join(testRootDir, tmpSubDataFolderName, "Dan Brown's The Lost Symbol - S01E03 - Murmuration WEBDL-1080p", "Dan Brown's The Lost Symbol - S01E03 - Murmuration WEBDL-1080p.chinese(inside).ass"),
			ch_enSubFile:           path.Join(testRootDir, tmpSubDataFolderName, "Dan Brown's The Lost Symbol - S01E03 - Murmuration WEBDL-1080p", "Dan Brown's The Lost Symbol - S01E03 - Murmuration WEBDL-1080p.chinese(繁英,xunlei).ass"),
			staticLineFileSavePath: "bar.html"},
			want: -0.505656, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bFind, infoBase, err := subParserHub.DetermineFileTypeFromFile(tt.args.enSubFile)
			if err != nil {
				t.Fatal(err)
			}
			if bFind == false {
				t.Fatal("sub not match")
			}
			/*
				这里发现一个梗，内置的英文字幕导出的时候，有可能需要合并多个 Dialogue，见
				internal/pkg/sub_helper/sub_helper.go 中 MergeMultiDialogue4EngSubtitle 的实现
			*/
			sub_helper.MergeMultiDialogue4EngSubtitle(infoBase)

			bFind, infoSrc, err := subParserHub.DetermineFileTypeFromFile(tt.args.ch_enSubFile)
			if err != nil {
				t.Fatal(err)
			}
			if bFind == false {
				t.Fatal("sub not match")
			}
			/*
				这里发现一个梗，内置的英文字幕导出的时候，有可能需要合并多个 Dialogue，见
				internal/pkg/sub_helper/sub_helper.go 中 MergeMultiDialogue4EngSubtitle 的实现
			*/
			sub_helper.MergeMultiDialogue4EngSubtitle(infoSrc)

			got, err := GetOffsetTime(infoBase, infoSrc, tt.args.ch_enSubFile+"-bar.html", tt.args.ch_enSubFile+".log")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOffsetTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 在一个正负范围内都可以接受
			if got > tt.want-0.1 && got < tt.want+0.1 {

			} else {
				t.Errorf("GetOffsetTime() got = %v, want %v", got, tt.want)
			}
			//if got != tt.want {
			//	t.Errorf("GetOffsetTime() got = %v, want %v", got, tt.want)
			//}
			println(fmt.Sprintf("GetOffsetTime: %fs", got))
		})
	}
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

const tmpSubDataFolderName = "SubFixCache"
