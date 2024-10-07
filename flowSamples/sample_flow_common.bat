@echo off
setlocal enabledelayedexpansion

set "local="
set "remote="
set "file="
set "param=_"
set "uparams="
set "chkingParam="
set "direction="
set "testTemp=testdata\flow.bat.msg"

goto main

:chkParam
set chkingParam=%~3
set uparams=%~4
if not "!chkingParam!" == "" (
	if not "!chkingParam!" == "_" (
		echo "!chkingParam! NEEDS a value!"
		set "uparams=!uparams! !chkingParam!"
		set "chkingParam="
	)
)
if not "!chkingParam!" == "" set "%~1=!chkingParam!"
set "%~2=!uparams!"
exit /b 0

:main
for %%x in (%*) do (
    	set i=%%x
	if "!direction!" == "" (
		set "direction=!i!"
	) else if "!i:~0,2!" equ "--" (
		call :chkParam chkingParam uparams "!param!" "!uparams!"
	    	set param=!i:~2!
    	) else if "!i:~0,1!" equ "-" (
		call :chkParam !param!
	    	set param=!i:~1!
    	) else if "!param!"=="local" (
        	set "local=!i!"
		set "param="
    	) else if "!param!"=="remote" (
        	set "remote=!i!"
		set "param="
    	) else if "!param!"=="file" (
        	set "file=!i!"
		set "param="
    	) else (
		call :chkParam chkingParam uparams "!param!" "!uparams!"
		if "!chkingParam!" == "" (
			goto nextloop
		)
		echo "!i! NOT recognized!"
		set "uparams=!uparams! !i!"
:nextloop
		set param=
	)
)
    if not "%param%" == "_" (
    	if not "%param%" == "" (
    		echo "!param! NOT recognized or NEEDS a value!"
    		set "uparams=!uparams! !param!"
    	)
    )
    echo local=%local%, remote=%remote%, file=%file%
    if "%file%"=="" goto fend
    if "%local%"=="" goto fend
    if "%direction%" == "in" (
        echo. >> %file%
        echo. IN: from %remote% to %local% >> %file%
        move %file% %testTemp%
    ) else if "%direction%" == "out" (
        echo. OUT: from %local% to %remote% >> %testTemp%
        move %testTemp% %file%
    ) else (
	echo "UNKNOWN caller!"
    )
    goto end
:fend
    if not "!file!"=="" echo unrecognized/incomplete params=!uparams! > !file!
    exit /b 1
:end
