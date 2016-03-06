// +build windows

package builder

func init() {
	echoScript = `..\fixtures\echo.bat`
	failScript = `..\fixtures\fail.bat`
}
