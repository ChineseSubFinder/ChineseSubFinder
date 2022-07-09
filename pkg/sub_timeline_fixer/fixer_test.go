package sub_timeline_fixer

import (
	"testing"

	"github.com/james-bowman/nlp"
	"github.com/james-bowman/nlp/measures/pairwise"
	"gonum.org/v1/gonum/mat"
)

func TestStopWordCounter(t *testing.T) {

	//testRootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"FixTimeline"}, 4, false)
	//subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())
	//bFind, info, err := subParserHub.DetermineFileTypeFromFile(filepath.Join(testRootDir, "org", "yes", "R&M S05E01 - English.srt"))
	//if err != nil {
	//	t.Fatal(err)
	//}
	//if bFind == false {
	//	t.Fatal("not match sub types")
	//}
	//
	//allString := strings.Join(info.OtherLines, " ")
	//
	//s := SubTimelineFixer{}
	//stopWords := s.StopWordCounter(strings.ToLower(allString), 5)
	//
	//t.Logf("\n\nsub name: %s \t lem(stopWords): %d", info.Name, len(stopWords))
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

	vectoriser := nlp.NewCountVectoriser(EnStopWords...)
	transformer := nlp.NewTfidfTransformer()

	// set k (the number of dimensions following truncation) to 4
	reducer := nlp.NewTruncatedSVD(4)

	lsiPipeline := nlp.NewPipeline(vectoriser, transformer, reducer)

	// Transform the corpus into an LSI fitting the model to the documents in the process
	lsi, err := lsiPipeline.FitTransform(testCorpus...)
	if err != nil {
		t.Errorf("Failed to process documents because %v", err)
		return
	}

	// run the query through the same pipeline that was fitted to the corpus and
	// to project it into the same dimensional space
	queryVector, err := lsiPipeline.Transform(query)
	if err != nil {
		t.Errorf("Failed to process documents because %v", err)
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

	t.Logf("\n\nMatched '%s'", testCorpus[matched])
	// Output: Matched 'The quick brown fox jumped over the lazy dog'
}

func TestGetOffsetTimeV1(t *testing.T) {

	//testRootDir := unit_test_helper.GetTestDataResourceRootPath([]string{"FixTimeline"}, 4, true)
	//
	//testRootDirYes := filepath.Join(testRootDir, "yes")
	//testRootDirNo := filepath.Join(testRootDir, "no")
	//subParserHub := sub_parser_hub.NewSubParserHub(ass.NewParser(), srt.NewParser())
	//
	//type args struct {
	//	enSubFile              string
	//	ch_enSubFile           string
	//	staticLineFileSavePath string
	//}
	//tests := []struct {
	//	name    string
	//	args    args
	//	want    float64
	//	wantErr bool
	//}{
	//	/*
	//		这里有几个比较理想的字幕时间轴校正的示例
	//	*/
	//	{name: "R&M S05E01", args: args{enSubFile: filepath.Join(testRootDirYes, "R&M S05E01 - English.srt"),
	//		ch_enSubFile:           filepath.Join(testRootDirYes, "R&M S05E01 - 简英.srt"),
	//		staticLineFileSavePath: "bar.html"}, want: -6.42981818181818, wantErr: false},
	//	{name: "R&M S05E10", args: args{enSubFile: filepath.Join(testRootDirYes, "R&M S05E10 - English.ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirYes, "R&M S05E10 - 简英.ass"),
	//		staticLineFileSavePath: "bar.html"}, want: -6.335985401459854, wantErr: false},
	//	{name: "基地 S01E03", args: args{enSubFile: filepath.Join(testRootDirYes, "基地 S01E03 - English.ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirYes, "基地 S01E03 - 简英.ass"),
	//		staticLineFileSavePath: "bar.html"}, want: -32.09061538461539, wantErr: false},
	//	/*
	//		WTF,这部剧集
	//		Dan Brown'timelineFixer The Lost Symbol
	//		内置的英文字幕时间轴是歪的，所以修正完了就错了
	//	*/
	//	{name: "Dan Brown'timelineFixer The Lost Symbol - S01E01", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E01.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E01.chinese(简英,shooter).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 1.3217821782178225, wantErr: false},
	//	{name: "Dan Brown'timelineFixer The Lost Symbol - S01E02", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E02.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E02.chinese(简英,subhd).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: -0.5253383458646617, wantErr: false},
	//	{name: "Dan Brown'timelineFixer The Lost Symbol - S01E03", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E03.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E03.chinese(繁英,xunlei).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: -0.505656, wantErr: false},
	//	{name: "Dan Brown'timelineFixer The Lost Symbol - S01E04", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E04.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Dan Brown's The Lost Symbol - S01E04.chinese(简英,zimuku).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: -0.633415, wantErr: false},
	//	/*
	//		只有一个是字幕下载了一个错误的，其他的无需修正
	//	*/
	//	{name: "Don't Breathe 2 (2021) - shooter-srt", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(inside).srt"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(简英,shooter).srt"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Don't Breathe 2 (2021) - subhd-srt error matched sub", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(inside).srt"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(简英,subhd).srt"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Don't Breathe 2 (2021) - xunlei-ass", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(简英,xunlei).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Don't Breathe 2 (2021) - zimuku-ass", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Don't Breathe 2 (2021).chinese(简英,zimuku).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	/*
	//		基地
	//	*/
	//	{name: "Foundation (2021) - S01E01", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Foundation (2021) - S01E01.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Foundation (2021) - S01E01.chinese(简英,zimuku).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Foundation (2021) - S01E02", args: args{
	//		enSubFile:              filepath.Join(testRootDirYes, "Foundation (2021) - S01E02.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirYes, "Foundation (2021) - S01E02.chinese(简英,subhd).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: -30.624840, wantErr: false},
	//	{name: "Foundation (2021) - S01E03", args: args{
	//		enSubFile:              filepath.Join(testRootDirYes, "Foundation (2021) - S01E03.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirYes, "Foundation (2021) - S01E03.chinese(简英,subhd).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: -32.085037037037054, wantErr: false},
	//	{name: "Foundation (2021) - S01E04", args: args{
	//		enSubFile:              filepath.Join(testRootDirYes, "Foundation (2021) - S01E04.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirYes, "Foundation (2021) - S01E04.chinese(简英,subhd).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: -36.885074, wantErr: false},
	//	{name: "Foundation (2021) - S01E04", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Foundation (2021) - S01E04.chinese(inside).srt"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Foundation (2021) - S01E04.chinese(繁英,shooter).srt"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	/*
	//		The Card Counter
	//	*/
	//	{name: "The Card Counter", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "The Card Counter (2021).chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "The Card Counter (2021).chinese(简英,xunlei).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "The Card Counter", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "The Card Counter (2021).chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "The Card Counter (2021).chinese(简英,shooter).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0.224844, wantErr: false},
	//	/*
	//		Kingdom Ashin of the North
	//	*/
	//	{name: "Kingdom Ashin of the North - error matched sub", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Kingdom Ashin of the North (2021).chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Kingdom Ashin of the North (2021).chinese(简英,subhd).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	/*
	//		Only Murders in the Building
	//	*/
	//	{name: "Only Murders in the Building - S01E06", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Only Murders in the Building - S01E06.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Only Murders in the Building - S01E06.chinese(简英,subhd).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Only Murders in the Building - S01E08", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Only Murders in the Building - S01E08.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Only Murders in the Building - S01E08.chinese(简英,subhd).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	/*
	//		Ted Lasso
	//	*/
	//	{name: "Ted Lasso - S02E09", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E09.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E09.chinese(简英,subhd).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Ted Lasso - S02E09", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E09.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E09.chinese(简英,zimuku).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Ted Lasso - S02E10", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(简英,subhd).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Ted Lasso - S02E10", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(简英,zimuku).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Ted Lasso - S02E10", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E10.chinese(简英,shooter).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Ted Lasso - S02E11", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E11.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E11.chinese(简英,subhd).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Ted Lasso - S02E11", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E11.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E11.chinese(简英,zimuku).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Ted Lasso - S02E12", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E12.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E12.chinese(简英,subhd).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "Ted Lasso - S02E12", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "Ted Lasso - S02E12.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "Ted Lasso - S02E12.chinese(简英,shooter).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	/*
	//		The Protégé
	//	*/
	//	{name: "The Protégé", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "The Protégé (2021).chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "The Protégé (2021).chinese(简英,zimuku).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "The Protégé", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "The Protégé (2021).chinese(inside).srt"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "The Protégé (2021).chinese(简英,shooter).srt"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	/*
	//		The Witcher Nightmare of the Wolf
	//	*/
	//	{name: "The Witcher Nightmare of the Wolf", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "The Witcher Nightmare of the Wolf.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "The Witcher Nightmare of the Wolf.chinese(简英,zimuku).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	/*
	//		What If…!
	//	*/
	//	{name: "What If…! - S01E07", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "What If…! - S01E07.chinese(inside).ass"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "What If…! - S01E07.chinese(简英,subhd).ass"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//	{name: "What If…! - S01E09", args: args{
	//		enSubFile:              filepath.Join(testRootDirNo, "What If…! - S01E09.chinese(inside).srt"),
	//		ch_enSubFile:           filepath.Join(testRootDirNo, "What If…! - S01E09.chinese(简英,shooter).srt"),
	//		staticLineFileSavePath: "bar.html"},
	//		want: 0, wantErr: false},
	//}
	//
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//
	//		bFind, infoBase, err := subParserHub.DetermineFileTypeFromFile(tt.args.enSubFile)
	//		if err != nil {
	//			t.Fatal(err)
	//		}
	//		if bFind == false {
	//			t.Fatal("sub not match")
	//		}
	//		/*
	//			这里发现一个梗，内置的英文字幕导出的时候，有可能需要合并多个 Dialogue，见
	//			internal/pkg/sub_helper/sub_helper.go 中 MergeMultiDialogue4EngSubtitle 的实现
	//		*/
	//		sub_helper.MergeMultiDialogue4EngSubtitle(infoBase)
	//
	//		bFind, infoSrc, err := subParserHub.DetermineFileTypeFromFile(tt.args.ch_enSubFile)
	//		if err != nil {
	//			t.Fatal(err)
	//		}
	//		if bFind == false {
	//			t.Fatal("sub not match")
	//		}
	//		/*
	//			这里发现一个梗，内置的英文字幕导出的时候，有可能需要合并多个 Dialogue，见
	//			internal/pkg/sub_helper/sub_helper.go 中 MergeMultiDialogue4EngSubtitle 的实现
	//		*/
	//		sub_helper.MergeMultiDialogue4EngSubtitle(infoSrc)
	//
	//		bok, got, sd, err := timelineFixer.GetOffsetTimeV1(infoBase, infoSrc, tt.args.ch_enSubFile+"-bar.html", tt.args.ch_enSubFile+".log")
	//		if (err != nil) != tt.wantErr {
	//			t.Errorf("GetOffsetTimeV1() error = %v, wantErr %v", err, tt.wantErr)
	//			return
	//		}
	//
	//		// 在一个正负范围内都可以接受
	//		if got > tt.want-0.1 && got < tt.want+0.1 {
	//
	//		} else {
	//			t.Errorf("GetOffsetTimeV1() got = %v, want %v", got, tt.want)
	//		}
	//		//if got != tt.want {
	//		//	t.Errorf("GetOffsetTimeV1() got = %v, want %v", got, tt.want)
	//		//}
	//
	//		if bok == true && got != 0 {
	//			_, err = timelineFixer.FixSubTimelineOneOffsetTime(infoSrc, got, tt.args.ch_enSubFile+FixMask+infoBase.Ext)
	//			if err != nil {
	//				t.Fatal(err)
	//			}
	//		}
	//
	//		println(fmt.Sprintf("GetOffsetTimeV1: %fs SD:%f", got, sd))
	//	})
	//}
}

//var timelineFixer = NewSubTimelineFixer(sub_timeline_fiexer.SubTimelineFixerConfig{
//	// V1
//	V1_MaxCompareDialogue: 3,
//	V1_MaxStartTimeDiffSD: 0.1,
//	V1_MinMatchedPercent:  0.1,
//	V1_MinOffset:          0.1,
//	// V2
//	V2_SubOneUnitProcessTimeOut: 5 * 60,
//	V2_FrontAndEndPerBase:       0.1,
//	V2_FrontAndEndPerSrc:        0.2,
//	V2_WindowMatchPer:           0.2,
//	V2_CompareParts:             3,
//	V2_FixThreads:               2,
//	V2_MaxStartTimeDiffSD:       0.1,
//	V2_MinOffset:                0.2,
//})
