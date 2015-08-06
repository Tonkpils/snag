// +build windows

package vow

func init() {
	echoScript = `..\fixtures\echo.bat`
	failScript = `..\fixtures\fail.bat`
}
