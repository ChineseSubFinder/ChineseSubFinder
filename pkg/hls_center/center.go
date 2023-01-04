package hls_center

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/ffmpeg_helper"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/hls_center/worker"
	"github.com/sirupsen/logrus"
	"io"
	"path/filepath"
	"text/template"
)

type Center struct {
	logger       *logrus.Logger
	ffmpegHelper *ffmpeg_helper.FFMPEGHelper
	encodeWorker *worker.WorkerServer
}

func NewCenter(logger *logrus.Logger) *Center {

	cacheRootDir, err := pkg.GetVideoAndSubPreviewCacheFolder()
	if err != nil {
		panic(err)
	}
	encodeWorker := worker.NewWorkerServer(worker.WorkerServerConf{
		NumWorkers: 2,
		CacheDir:   filepath.Join(cacheRootDir, "segments"),
		Worker:     worker.NewCommandWorker("ffmpeg"),
	})

	return &Center{
		logger:       logger,
		ffmpegHelper: ffmpeg_helper.NewFFMPEGHelper(logger),
		encodeWorker: encodeWorker,
	}
}

// WritePlaylist 构建 m3u8 文件
func (c *Center) WritePlaylist(urlTemplate string, videoFileFPath string, w io.Writer) error {

	t := template.Must(template.New("urlTemplate").Parse(urlTemplate))

	if pkg.IsFile(videoFileFPath) == false {
		return errors.New("WritePlaylist video file not exist or it's blu-ray, not support yet, file = " + videoFileFPath)
	}

	duration := c.ffmpegHelper.GetVideoDuration(videoFileFPath)
	getUrl := func(segmentIndex int) string {
		buf := new(bytes.Buffer)
		t.Execute(buf, struct {
			Resolution int64
			Segment    int
		}{
			720,
			segmentIndex,
		})
		return buf.String()
	}

	fmt.Fprint(w, "#EXTM3U\n")
	fmt.Fprint(w, "#EXT-X-VERSION:3\n")
	fmt.Fprint(w, "#EXT-X-MEDIA-SEQUENCE:0\n")
	fmt.Fprint(w, "#EXT-X-ALLOW-CACHE:YES\n")
	fmt.Fprint(w, "#EXT-X-TARGETDURATION:"+fmt.Sprintf("%.f", hlsSegmentLength)+"\n")
	fmt.Fprint(w, "#EXT-X-PLAYLIST-TYPE:VOD\n")

	leftover := duration
	segmentIndex := 0

	for leftover > 0 {
		if leftover > hlsSegmentLength {
			fmt.Fprintf(w, "#EXTINF: %f,\n", hlsSegmentLength)
		} else {
			fmt.Fprintf(w, "#EXTINF: %f,\n", leftover)
		}
		fmt.Fprintf(w, getUrl(segmentIndex)+"\n")
		segmentIndex++
		leftover = leftover - hlsSegmentLength
	}
	fmt.Fprint(w, "#EXT-X-ENDLIST\n")
	return nil
}

// WriteSegment 构建 ts 文件
func (c *Center) WriteSegment(videoFileFPath string, segmentIndex int64, resolution int64, w io.Writer) error {

	if pkg.IsFile(videoFileFPath) == false {
		return errors.New("WriteSegment video file not exist, file = " + videoFileFPath)
	}

	args := encodingArgs(videoFileFPath, segmentIndex, resolution)
	return c.encodeWorker.Serve(args, w)
}

func encodingArgs(videoFile string, segment int64, resolution int64) []string {
	startTime := segment * hlsSegmentLength
	// see http://superuser.com/questions/908280/what-is-the-correct-way-to-fix-keyframes-in-ffmpeg-for-dash
	return []string{
		// Prevent encoding to run longer than 30 seonds
		"-timelimit", "45",

		// TODO: Some stuff to investigate
		// "-probesize", "524288",
		// "-fpsprobesize", "10",
		// "-analyzeduration", "2147483647",
		// "-hwaccel:0", "vda",

		// The start time
		// important: needs to be before -i to do input seeking
		"-ss", fmt.Sprintf("%v.00", startTime),

		// The source file
		"-i", videoFile,

		// Put all streams to output
		// "-map", "0",

		// The duration
		"-t", fmt.Sprintf("%v.00", hlsSegmentLength),

		// TODO: Find out what it does
		//"-strict", "-2",

		// Synchronize audio
		"-async", "1",

		// 720p
		"-vf", fmt.Sprintf("scale=-2:%v", resolution),

		// x264 video codec
		"-vcodec", "libx264",

		// x264 preset
		"-preset", "veryfast",

		// aac audio codec
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "2",

		// TODO
		"-pix_fmt", "yuv420p",

		//"-r", "25", // fixed framerate

		"-force_key_frames", "expr:gte(t,n_forced*5.000)",

		//"-force_key_frames", "00:00:00.00",
		//"-x264opts", "keyint=25:min-keyint=25:scenecut=-1",

		//"-f", "mpegts",

		"-f", "ssegment",
		"-segment_time", fmt.Sprintf("%v.00", hlsSegmentLength),
		"-initial_offset", fmt.Sprintf("%v.00", startTime),

		"pipe:out%03d.ts",
	}
}

const hlsSegmentLength = 5.0 // Seconds
