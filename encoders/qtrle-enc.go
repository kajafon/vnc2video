package encoders

import (
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/amitbet/vnc2video/logger"
)

// QTRLEImageEncoder quick time rle is an efficient loseless codec, uses .mov extension
type QTRLEImageEncoder struct {
	FFMpegBinPath string
	cmd           *exec.Cmd
	input         io.WriteCloser
	closed        bool
	Framerate     int
}

func (enc *QTRLEImageEncoder) Init(videoFileName string, output io.Writer) {
	fileExt := ".mov"
	if enc.Framerate == 0 {
		enc.Framerate = 12
	}
	if !strings.HasSuffix(videoFileName, fileExt) {
		videoFileName = videoFileName + fileExt
	}
	//binary := "./ffmpeg"
	cmd := exec.Command(enc.FFMpegBinPath,
		"-f", "image2pipe",
		"-vcodec", "ppm",
		//"-r", strconv.Itoa(framerate),
		"-r", fmt.Sprint(enc.Framerate),

		//"-re",
		//"-i", "pipe:0",
		"-an", //no audio
		//"-vsync", "2",
		///"-probesize", "10000000",
		"-y",

		"-i", "-",
		// "–size", "1920x1080",
		"-vcodec", "qtrle", //"libvpx",//"libvpx-vp9"//"libx264"
		//"-b:v", "0.33M",
		"-threads", "7",
		"-preset", "veryfast",
		///"-coder", "1",
		///"-bf", "0",
		///"-me_method", "hex",
		//"-speed", "0",
		//"-lossless", "1", //for vpx
		// "-an", "-f", "webm",
		//"-tune", "animation",
		"-maxrate", "0.5M",
		"-bufsize", "50M",
		"-g", "250",

		//"-crf", "0", //for lossless encoding!!!!

		//"-rc_lookahead", "16",
		//"-profile", "0",
		"-crf", "34",
		//"-qmax", "51",
		//"-qmin", "7",
		//"-slices", "4",
		//"-vb", "2M",

		videoFileName,
	)
	//cmd := exec.Command("/bin/echo")

	//io.Copy(cmd.Stdout, os.Stdout)
	if output != nil {
		cmd.Stdout = output //os.Stdout
		cmd.Stderr = output // os.Stderr
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	encInput, err := cmd.StdinPipe()
	enc.input = encInput
	if err != nil {
		logger.Error("can't get ffmpeg input pipe")
	}
	enc.cmd = cmd
}
func (enc *QTRLEImageEncoder) Run(videoFileName string, output io.Writer) error {
	// if _, err := os.Stat(enc.FFMpegBinPath); os.IsNotExist(err) {
	// 	if _, err := os.Stat(enc.FFMpegBinPath + ".exe"); os.IsNotExist(err) {
	// 		logger.Error("encoder file doesn't exist in path:", enc.FFMpegBinPath)
	// 		return errors.New("encoder file doesn't exist in path" + videoFileName)
	// 	} else {
	// 		enc.FFMpegBinPath = enc.FFMpegBinPath + ".exe"
	// 	}
	// }

	enc.Init(videoFileName, output)
	logger.Debugf("launching binary: %v", enc.cmd)
	err := enc.cmd.Run()
	if err != nil {
		logger.Errorf("error while launching ffmpeg: %v\n err: %v", enc.cmd.Args, err)
		return err
	}
	return nil
}
func (enc *QTRLEImageEncoder) Encode(img image.Image) error {
	if enc.input == nil || enc.closed {
		return nil
	}

	err := encodePPM(enc.input, img)
	if err != nil {
		logger.Error("error while encoding image:", err)
	}
	return err
}

func (enc *QTRLEImageEncoder) Close() {
	enc.closed = true
	if enc.cmd != nil && enc.cmd.Process != nil {
		enc.cmd.Process.Signal(os.Interrupt)
	}
	if enc.input != nil {
		enc.input.Close()
	}
}
