#!/bin/sh
current=5
cp -va ../NOTICE                  config/win/
cp -va resources/$current.ico     config/win/systray.ico
cp -va ../NOTICE                  packaging/win/
cp -va resources/$current.ico     packaging/win/exe.ico
cp -va ../NOTICE                  config/mac/
cp -va resources/$current.icns    config/mac/menubar.icns
cp -va ../NOTICE                  packaging/mac/Resources/
cp -va resources/$current.icns    packaging/mac/Resources/app.icns