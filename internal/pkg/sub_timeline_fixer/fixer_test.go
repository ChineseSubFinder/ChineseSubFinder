package sub_timeline_fixer

import (
	"fmt"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/ass"
	"github.com/allanpk716/ChineseSubFinder/internal/logic/sub_parser/srt"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/debug_view"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/my_util"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_helper"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/sub_parser_hub"
	"github.com/allanpk716/ChineseSubFinder/internal/pkg/vad"
	"github.com/allanpk716/ChineseSubFinder/internal/types/sub_timeline_fiexer"
	"github.com/james-bowman/nlp"
	"github.com/james-bowman/nlp/measures/pairwise"
	"gonum.org/v1/gonum/mat"
	"path/filepath"
	"strings"
	"testing"
)

func TestStopWordCounter(t *testing.T) {

	testDataPath := "../../../TestData/FixTimeline"
	testRootDir, err := my_util.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}

	subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())

	bFind, info, err := subParserHub.DetermineFileTypeFromFile(filepath.Join(testRootDir, "R&M S05E10 - English.srt"))
	if err != nil {
		t.Fatal(err)
	}
	if bFind == false {
		t.Fatal("not match sub types")
	}

	allString := strings.Join(info.OtherLines, " ")

	s := SubTimelineFixer{}
	stopWords := s.StopWordCounter(strings.ToLower(allString), 5)

	print(len(stopWords))
	println(info.Name)
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

func TestGetOffsetTimeV1(t *testing.T) {
	testDataPath := "../../../TestData/FixTimeline"
	testRootDir, err := my_util.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}
	testRootDirYes := filepath.Join(testRootDir, "yes")
	testRootDirNo := filepath.Join(testRootDir, "no")
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
		/*
			这里有几个比较理想的字幕时间轴校正的示例
		*/
		{name: "R&M S05E01", args: args{enSubFile: filepath.Join(testRootDirYes, "R&M S05E01 - English.srt"),
			ch_enSubFile:           filepath.Join(testRootDirYes, "R&M S05E01 - 简英.srt"),
			staticLineFileSavePath: "bar.html"}, want: -6.42981818181818, wantErr: false},
		{name: "R&M S05E10", args: args{enSubFile: filepath.Join(testRootDirYes, "R&M S05E10 - English.ass"),
			ch_enSubFile:           filepath.Join(testRootDirYes, "R&M S05E10 - 简英.ass"),
			staticLineFileSavePath: "bar.html"}, want: -6.335985401459854, wantErr: false},
		{name: "基地 S01E03", args: args{enSubFile: filepath.Join(testRootDirYes, "基地 S01E03 - English.ass"),
			ch_enSubFile:           filepath.Join(testRootDirYes, "基地 S01E03 - 简英.ass"),
			staticLineFileSavePath: "bar.html"}, want: -32.09061538461539, wantErr: false},
		/*
			WTF,这部剧集
			Dan Brown'timelineFixer The Lost Symbol
			内置的英文字幕时间轴是歪的，所以修正完了就错了
		*/
		{name: "Dan Brown'timelineFixer The Lost Symbol - S01E01", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E01.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E01.chinese(简英,shooter).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 1.3217821782178225, wantErr: false},
		{name: "Dan Brown'timelineFixer The Lost Symbol - S01E02", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E02.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E02.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: -0.5253383458646617, wantErr: false},
		{name: "Dan Brown'timelineFixer The Lost Symbol - S01E03", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E03.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E03.chinese(繁英,xunlei).ass"),
			staticLineFileSavePath: "bar.html"},
			want: -0.505656, wantErr: false},
		{name: "Dan Brown'timelineFixer The Lost Symbol - S01E04", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E04.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E04.chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: -0.633415, wantErr: false},
		/*
			只有一个是字幕下载了一个错误的，其他的无需修正
		*/
		{name: "Don't Breathe 2 (2021) - shooter-srt", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(inside).srt"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(简英,shooter).srt"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Don't Breathe 2 (2021) - subhd-srt error matched sub", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(inside).srt"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(简英,subhd).srt"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Don't Breathe 2 (2021) - xunlei-ass", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(简英,xunlei).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Don't Breathe 2 (2021) - zimuku-ass", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			基地
		*/
		{name: "Foundation (2021) - S01E01", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Foundation (2021) - S01E01.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Foundation (2021) - S01E01.chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Foundation (2021) - S01E02", args: args{
			enSubFile:              filepath.Join(testRootDirYes, "Foundation (2021) - S01E02.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirYes, "Foundation (2021) - S01E02.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: -30.624840, wantErr: false},
		{name: "Foundation (2021) - S01E03", args: args{
			enSubFile:              filepath.Join(testRootDirYes, "Foundation (2021) - S01E03.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirYes, "Foundation (2021) - S01E03.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: -32.085037037037054, wantErr: false},
		{name: "Foundation (2021) - S01E04", args: args{
			enSubFile:              filepath.Join(testRootDirYes, "Foundation (2021) - S01E04.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirYes, "Foundation (2021) - S01E04.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: -36.885074, wantErr: false},
		{name: "Foundation (2021) - S01E04", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Foundation (2021) - S01E04.chinese(inside).srt"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Foundation (2021) - S01E04.chinese(繁英,shooter).srt"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			The Card Counter
		*/
		{name: "The Card Counter", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "The Card Counter (2021).chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "The Card Counter (2021).chinese(简英,xunlei).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "The Card Counter", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "The Card Counter (2021).chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "The Card Counter (2021).chinese(简英,shooter).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0.224844, wantErr: false},
		/*
			Kingdom Ashin of the North
		*/
		{name: "Kingdom Ashin of the North - error matched sub", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Kingdom Ashin of the North (2021).chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Kingdom Ashin of the North (2021).chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			Only Murders in the Building
		*/
		{name: "Only Murders in the Building - S01E06", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Only Murders in the Building - S01E06.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Only Murders in the Building - S01E06.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Only Murders in the Building - S01E08", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Only Murders in the Building - S01E08.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Only Murders in the Building - S01E08.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			Ted Lasso
		*/
		{name: "Ted Lasso - S02E09", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E09.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E09.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E09", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E09.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E09.chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E10", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E10", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E10", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(简英,shooter).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E11", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E11.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E11.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E11", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E11.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E11.chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E12", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E12.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E12.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E12", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E12.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E12.chinese(简英,shooter).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			The Protégé
		*/
		{name: "The Protégé", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "The Protégé (2021).chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "The Protégé (2021).chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "The Protégé", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "The Protégé (2021).chinese(inside).srt"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "The Protégé (2021).chinese(简英,shooter).srt"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			The Witcher Nightmare of the Wolf
		*/
		{name: "The Witcher Nightmare of the Wolf", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "The Witcher Nightmare of the Wolf.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "The Witcher Nightmare of the Wolf.chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			What If…!
		*/
		{name: "What If…! - S01E07", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "What If…! - S01E07.chinese(inside).ass"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "What If…! - S01E07.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "What If…! - S01E09", args: args{
			enSubFile:              filepath.Join(testRootDirNo, "What If…! - S01E09.chinese(inside).srt"),
			ch_enSubFile:           filepath.Join(testRootDirNo, "What If…! - S01E09.chinese(简英,shooter).srt"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
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

			bok, got, sd, err := timelineFixer.GetOffsetTimeV1(infoBase, infoSrc, tt.args.ch_enSubFile+"-bar.html", tt.args.ch_enSubFile+".log")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOffsetTimeV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 在一个正负范围内都可以接受
			if got > tt.want-0.1 && got < tt.want+0.1 {

			} else {
				t.Errorf("GetOffsetTimeV1() got = %v, want %v", got, tt.want)
			}
			//if got != tt.want {
			//	t.Errorf("GetOffsetTimeV1() got = %v, want %v", got, tt.want)
			//}

			if bok == true && got != 0 {
				_, err = timelineFixer.FixSubTimelineOneOffsetTime(infoSrc, got, tt.args.ch_enSubFile+FixMask+infoBase.Ext)
				if err != nil {
					t.Fatal(err)
				}
			}

			println(fmt.Sprintf("GetOffsetTimeV1: %fs SD:%f", got, sd))
		})
	}
}

func TestGetOffsetTimeV2_BaseSub(t *testing.T) {
	testDataPath := "../../../TestData/FixTimeline"
	testRootDir, err := my_util.CopyTestData(testDataPath)
	if err != nil {
		t.Fatal(err)
	}
	testRootDirYes := filepath.Join(testRootDir, "yes")
	testRootDirNo := filepath.Join(testRootDir, "no")
	subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())

	type args struct {
		baseSubFile            string
		srcSubFile             string
		staticLineFileSavePath string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		/*
			这里有几个比较理想的字幕时间轴校正的示例
		*/
		{name: "R&M S05E01", args: args{baseSubFile: filepath.Join(testRootDirYes, "R&M S05E01 - English.srt"),
			srcSubFile:             filepath.Join(testRootDirYes, "R&M S05E01 - 简英.srt"),
			staticLineFileSavePath: "bar.html"}, want: -6.4, wantErr: false},
		{name: "R&M S05E01-1", args: args{baseSubFile: filepath.Join(testRootDirYes, "R&M S05E01 - English.srt"),
			srcSubFile:             filepath.Join(testRootDirYes, "R&M S05E01 - English.srt"),
			staticLineFileSavePath: "bar.html"}, want: 0, wantErr: false},

		{name: "R&M S05E10-0", args: args{baseSubFile: filepath.Join(testRootDirYes, "R&M S05E10 - English.ass"),
			srcSubFile:             filepath.Join(testRootDirYes, "R&M S05E10 - 简英.ass"),
			staticLineFileSavePath: "bar.html"}, want: -6.405985401459854, wantErr: false},

		{name: "R&M S05E10-1", args: args{baseSubFile: filepath.Join(testRootDirYes, "R&M S05E10 - 简英.ass"),
			srcSubFile:             filepath.Join(testRootDirYes, "R&M S05E10 - English.ass"),
			staticLineFileSavePath: "bar.html"}, want: 6.405985401459854, wantErr: false},
		{name: "R&M S05E10-2", args: args{baseSubFile: filepath.Join(testRootDirYes, "R&M S05E10 - 简英.ass"),
			srcSubFile:             filepath.Join(testRootDirYes, "R&M S05E10 - 简英.ass"),
			staticLineFileSavePath: "bar.html"}, want: 0, wantErr: false},
		/*
			基地
		*/
		{name: "Foundation (2021) - S01E01", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Foundation (2021) - S01E01.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Foundation (2021) - S01E01.chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Foundation (2021) - S01E02", args: args{
			baseSubFile:            filepath.Join(testRootDirYes, "Foundation (2021) - S01E02.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirYes, "Foundation (2021) - S01E02.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: -30.624840, wantErr: false},
		{name: "Foundation (2021) - S01E03", args: args{
			baseSubFile:            filepath.Join(testRootDirYes, "Foundation (2021) - S01E03.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirYes, "Foundation (2021) - S01E03.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: -32.085037037037054, wantErr: false},
		{name: "Foundation (2021) - S01E04", args: args{
			baseSubFile:            filepath.Join(testRootDirYes, "Foundation (2021) - S01E04.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirYes, "Foundation (2021) - S01E04.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: -36.885074, wantErr: false},
		{name: "Foundation (2021) - S01E04", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Foundation (2021) - S01E04.chinese(inside).srt"),
			srcSubFile:             filepath.Join(testRootDirNo, "Foundation (2021) - S01E04.chinese(繁英,shooter).srt"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			Don't Breathe 2 (2021)
		*/
		{name: "Don't Breathe 2 (2021) - zimuku-ass", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Don't Breathe 2 (2021) - shooter-srt", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(inside).srt"),
			srcSubFile:             filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(简英,shooter).srt"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			Only Murders in the Building
		*/
		{name: "Only Murders in the Building - S01E06", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Only Murders in the Building - S01E06.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Only Murders in the Building - S01E06.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Only Murders in the Building - S01E08", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Only Murders in the Building - S01E08.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Only Murders in the Building - S01E08.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			Ted Lasso
		*/
		{name: "Ted Lasso - S02E09", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Ted Lasso - S02E09.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Ted Lasso - S02E09.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E09", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Ted Lasso - S02E09.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Ted Lasso - S02E09.chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E10", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E10", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E10", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(简英,shooter).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E11", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Ted Lasso - S02E11.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Ted Lasso - S02E11.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E11", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Ted Lasso - S02E11.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Ted Lasso - S02E11.chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E12", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Ted Lasso - S02E12.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Ted Lasso - S02E12.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "Ted Lasso - S02E12", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "Ted Lasso - S02E12.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "Ted Lasso - S02E12.chinese(简英,shooter).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			The Protégé
		*/
		{name: "The Protégé", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "The Protégé (2021).chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "The Protégé (2021).chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "The Protégé", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "The Protégé (2021).chinese(inside).srt"),
			srcSubFile:             filepath.Join(testRootDirNo, "The Protégé (2021).chinese(简英,shooter).srt"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			The Witcher Nightmare of the Wolf
		*/
		{name: "The Witcher Nightmare of the Wolf", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "The Witcher Nightmare of the Wolf.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "The Witcher Nightmare of the Wolf.chinese(简英,zimuku).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		/*
			What If…!
		*/
		{name: "What If…! - S01E01", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "What If…! - S01E01_英_2.srt"),
			srcSubFile:             filepath.Join(testRootDirNo, "What If…! - S01E01 - (简英,shooter).srt"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "What If…! - S01E07", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "What If…! - S01E07.chinese(inside).ass"),
			srcSubFile:             filepath.Join(testRootDirNo, "What If…! - S01E07.chinese(简英,subhd).ass"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
		{name: "What If…! - S01E09", args: args{
			baseSubFile:            filepath.Join(testRootDirNo, "What If…! - S01E09.chinese(inside).srt"),
			srcSubFile:             filepath.Join(testRootDirNo, "What If…! - S01E09.chinese(简英,shooter).srt"),
			staticLineFileSavePath: "bar.html"},
			want: 0, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bFind, infoBase, err := subParserHub.DetermineFileTypeFromFile(tt.args.baseSubFile)
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
			//sub_helper.MergeMultiDialogue4EngSubtitle(infoBase)

			bFind, infoSrc, err := subParserHub.DetermineFileTypeFromFile(tt.args.srcSubFile)
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
			//sub_helper.MergeMultiDialogue4EngSubtitle(infoSrc)
			// ---------------------------------------------------------------------------------------
			// Base，截取的部分要大于 Src 的部分
			//baseUnitListOld, err := sub_helper.GetVADInfoFeatureFromSub(infoBase, V2_FrontAndEndPerBase, 100000, true)
			//if err != nil {
			//	t.Fatal(err)
			//}
			//baseUnitOld := baseUnitListOld[0]
			baseUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoBase, timelineFixer.FixerConfig.V2_FrontAndEndPerBase)
			if err != nil {
				t.Fatal(err)
			}
			// ---------------------------------------------------------------------------------------
			// Src，截取的部分要小于 Base 的部分
			//srcUnitListOld, err := sub_helper.GetVADInfoFeatureFromSub(infoSrc, V2_FrontAndEndPerSrc, 100000, true)
			//if err != nil {
			//	t.Fatal(err)
			//}
			//srcUnitOld := srcUnitListOld[0]
			srcUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoSrc, timelineFixer.FixerConfig.V2_FrontAndEndPerSrc)
			if err != nil {
				t.Fatal(err)
			}
			// ---------------------------------------------------------------------------------------
			//bok, got, sd, err := timelineFixer.GetOffsetTimeV2(&baseUnitOld, &srcUnitOld, nil, 0)
			bok, _, err := timelineFixer.GetOffsetTimeV2(baseUnitNew, srcUnitNew, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOffsetTimeV2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if bok == false {
				t.Fatal("GetOffsetTimeV2 return false")
			}

			//if got > -0.2 && got < 0.2 && tt.want == 0 {
			//	// 如果 offset time > -0.2 且 < 0.2 则认为无需调整时间轴，为0
			//} else if got > tt.want-0.1 && got < tt.want+0.1 {
			//	// 在一个正负范围内都可以接受
			//} else {
			//	t.Errorf("GetOffsetTimeV2() got = %v, want %v", got, tt.want)
			//}

			debug_view.SaveDebugChart(*baseUnitNew, tt.name+" -- baseUnitNew", "baseUnitNew")
			debug_view.SaveDebugChart(*srcUnitNew, tt.name+" -- srcUnitNew", "srcUnitNew")

			//if bok == true && got != 0 {
			//_, err = timelineFixer.FixSubTimelineOneOffsetTime(infoSrc, got, tt.args.srcSubFile+FixMask+infoBase.Ext)
			//if err != nil {
			//	t.Fatal(err)
			//}
			////}
			//println(fmt.Sprintf("GetOffsetTimeV2: %fs SD:%f", got, sd))
		})
	}
}

func TestGetOffsetTimeV2_BaseAudio(t *testing.T) {

	subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())

	type fields struct {
		fixerConfig sub_timeline_fiexer.SubTimelineFixerConfig
	}
	type args struct {
		audioInfo   vad.AudioInfo
		subFilePath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		want1   float64
		want2   float64
		wantErr bool
	}{
		// Rick and Morty - S05E01
		{name: "Rick and Morty - S05E01 -- 0",
			args: args{audioInfo: vad.AudioInfo{
				FileFullPath: "C:\\Tmp\\Rick and Morty - S05E01\\未知语言_1.pcm"},
				subFilePath: "C:\\Tmp\\Rick and Morty - S05E01\\英_2.ass"},
			want: true, want1: 0,
		},
		{name: "Rick and Morty - S05E01 -- 0",
			args: args{audioInfo: vad.AudioInfo{
				FileFullPath: "C:\\WorkSpace\\Go2hell\\src\\github.com\\allanpk716\\ChineseSubFinder\\internal\\pkg\\ffmpeg_helper\\CSF-SubFixCache\\Blade Runner - Black Lotus - S01E03 - The Human Condition WEBDL-1080p\\日_1.pcm"},
				subFilePath: "C:\\WorkSpace\\Go2hell\\src\\github.com\\allanpk716\\ChineseSubFinder\\CSF-SubFixCache\\Blade Runner - Black Lotus - S01E03 - The Human Condition WEBDL-1080p\\tar.ass"},
			want: true, want1: -5.1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bFind, infoSrc, err := subParserHub.DetermineFileTypeFromFile(tt.args.subFilePath)
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
			//sub_helper.MergeMultiDialogue4EngSubtitle(infoSrc)
			// Src，截取的部分要小于 Base 的部分
			srcUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoSrc, timelineFixer.FixerConfig.V2_FrontAndEndPerSrc)
			if err != nil {
				t.Fatal(err)
			}
			audioVADInfos, err := vad.GetVADInfoFromAudio(vad.AudioInfo{
				FileFullPath: tt.args.audioInfo.FileFullPath,
				SampleRate:   16000,
				BitDepth:     16,
			}, true)
			if err != nil {
				t.Fatal(err)
			}

			println("-------New--------")
			bok, _, err := timelineFixer.GetOffsetTimeV2(nil, srcUnitNew, audioVADInfos)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOffsetTimeV2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			//debug_view.SaveDebugChartBase(audioVADInfos, tt.name+" audioVADInfos", "audioVADInfos")
			//debug_view.SaveDebugChart(*srcUnitNew, tt.name+" srcUnitNew", "srcUnitNew")
			if bok != tt.want {
				t.Errorf("GetOffsetTimeV2() bok = %v, want %v", bok, tt.want)
			}

			//if offsetTime > -0.2 && offsetTime < 0.2 && tt.want1 == 0 {
			//	// 如果 offset time > -0.2 且 < 0.2 则认为无需调整时间轴，为0
			//} else if offsetTime > tt.want1-0.1 && offsetTime < tt.want1+0.1 {
			//	// 在一个正负范围内都可以接受
			//} else {
			//	t.Errorf("GetOffsetTimeV2() bok = %v, want %v", offsetTime, tt.want1)
			//}

			//_, err = timelineFixer.FixSubTimelineOneOffsetTime(infoSrc, offsetTime, tt.args.subFilePath+FixMask+infoSrc.Ext)
			//if err != nil {
			//	t.Fatal(err)
			//}

			//println(fmt.Sprintf("GetOffsetTimeV2: %vs SD:%v", offsetTime, sd))
		})
	}
}

func TestGetOffsetTimeV2_MoreTest(t *testing.T) {
	subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())

	type args struct {
		baseSubFile string
		srcSubFile  string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{name: "BL S01E03", args: args{
			baseSubFile: "C:\\Tmp\\BL - S01E03\\英_2.ass",
			srcSubFile:  "C:\\Tmp\\BL - S01E03\\org.ass",
		}, want: -4.1, wantErr: false},
		{name: "Rick and Morty - S05E10", args: args{
			baseSubFile: "C:\\Tmp\\Rick and Morty - S05E10\\英_2.ass",
			srcSubFile:  "C:\\Tmp\\Rick and Morty - S05E10\\org.ass",
		}, want: -4.1, wantErr: false},
		{name: "mix", args: args{
			baseSubFile: "C:\\Tmp\\Rick and Morty - S05E10\\英_2.ass",
			srcSubFile:  "C:\\Tmp\\BL - S01E03\\org.ass",
		}, want: -4.1, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bFind, infoBase, err := subParserHub.DetermineFileTypeFromFile(tt.args.baseSubFile)
			if err != nil {
				t.Fatal(err)
			}
			if bFind == false {
				t.Fatal("sub not match")
			}

			bFind, infoSrc, err := subParserHub.DetermineFileTypeFromFile(tt.args.srcSubFile)
			if err != nil {
				t.Fatal(err)
			}
			if bFind == false {
				t.Fatal("sub not match")
			}
			// ---------------------------------------------------------------------------------------
			// Base，截取的部分要大于 Src 的部分
			baseUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoBase, timelineFixer.FixerConfig.V2_FrontAndEndPerBase)
			if err != nil {
				t.Fatal(err)
			}
			// ---------------------------------------------------------------------------------------
			// Src，截取的部分要小于 Base 的部分
			srcUnitNew, err := sub_helper.GetVADInfoFeatureFromSubNew(infoSrc, timelineFixer.FixerConfig.V2_FrontAndEndPerSrc)
			if err != nil {
				t.Fatal(err)
			}
			// ---------------------------------------------------------------------------------------

			//bok, got, sd, err := timelineFixer.GetOffsetTimeV1(infoBase, infoSrc, tt.args.srcSubFile+"-bar.html", tt.args.srcSubFile+".log")
			//if (err != nil) != tt.wantErr {
			//	t.Errorf("GetOffsetTimeV1() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			bok, fixedResults, err := timelineFixer.GetOffsetTimeV2(baseUnitNew, srcUnitNew, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOffsetTimeV2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if bok == false {
				t.Fatal("GetOffsetTimeV2 return false")
			}

			debug_view.SaveDebugChart(*baseUnitNew, tt.name+" -- baseUnitNew", "baseUnitNew")
			debug_view.SaveDebugChart(*srcUnitNew, tt.name+" -- srcUnitNew", "srcUnitNew")

			_, err = timelineFixer.FixSubTimelineByFixResults(infoSrc, srcUnitNew, fixedResults, tt.args.srcSubFile+FixMask+infoBase.Ext)
			if err != nil {
				t.Fatal(err)
			}

		})
	}
}

var timelineFixer = NewSubTimelineFixer(sub_timeline_fiexer.SubTimelineFixerConfig{
	// V1
	V1_MaxCompareDialogue: 3,
	V1_MaxStartTimeDiffSD: 0.1,
	V1_MinMatchedPercent:  0.1,
	V1_MinOffset:          0.1,
	// V2
	V2_SubOneUnitProcessTimeOut: 5 * 60,
	V2_FrontAndEndPerBase:       0.1,
	V2_FrontAndEndPerSrc:        0.2,
	V2_WindowMatchPer:           0.2,
	V2_CompareParts:             3,
	V2_FixThreads:               2,
	V2_MaxStartTimeDiffSD:       0.1,
	V2_MinOffset:                0.2,
})
