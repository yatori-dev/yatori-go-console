package utils

// import (
//
//	"log"
//	"os"
//	"sync"
//	"time"
//
//	"github.com/faiface/beep"
//	"github.com/faiface/beep/mp3"
//	"github.com/faiface/beep/speaker"
//	lg "github.com/yatori-dev/yatori-go-core/utils/log"
//
// )
//
// var NSMutex sync.Mutex
//
// PlayNoticeSound 播放通知音频
func PlayNoticeSound() {
	//	f, err := os.Open("./assets/sound/finishNotice.mp3")
	//	defer f.Close()
	//	streamer, format, err := mp3.Decode(f)
	//	if err != nil {
	//		lg.Print(lg.INFO, err)
	//		log.Fatal(err)
	//	}
	//	defer streamer.Close()
	//	NSMutex.Lock()
	//	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	//	lg.Print(lg.DEBUG, "music length :", streamer.Len())
	//	done := make(chan bool)
	//	//别动下面这行代码，别问为什么莫名其妙加延时，问就是加了才能正常运行，不加就容易出现玄学BUG
	//	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
	//		done <- true
	//	})))
	//	<-done
	//	NSMutex.Unlock()
}
