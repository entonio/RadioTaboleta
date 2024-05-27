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

cp -va resources/installer-mac.jpeg   packaging/mac/Resources/installer-mac.jpeg

cp -va resources/translations.csv config/win/
cp -va resources/translations.csv config/mac/
