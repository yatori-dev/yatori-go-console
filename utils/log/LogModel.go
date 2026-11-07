package log

import lg "github.com/yatori-dev/yatori-go-core/v3/utils/log"

func ModelPrint(isSw bool, LOGLEVEL lg.LOGLEVEL, logText ...interface{}) {
	if isSw {
		lg.Print(LOGLEVEL, logText...)
	}
}
