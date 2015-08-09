// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
This repository hosts source code for mobile godoc app, for both Android and iOS.

This app is a typical example where Go is used for the main logic and
mobile platform's SDK language (Java for Android, Objective-C for iOS) is
used for UI elements. Language binding between Go and the SDK language is
generated with the gomobile bind tool. For more technical details on language
binding generation, see https://golang.org/x/mobile/cmd/gobind.

The web server code and the HTML contents in this app are those used by
https://golang.org and the Godoc command (https://golang.org/x/tools/cmd/godoc).
The current UI is a WebView, where URLs for golang.org are served by the
web server locally running. This app itself does not access the network.
When a user clicks a link that cannot be served by the local web server,
the app launches another App (the default browser) to handle the URL.

Build instruction

* Android

0. Download and initialize the gomobile tool. The gomobile tool requires
Go 1.5 or newer versions.

   > go get golang.org/x/mobile/cmd/gomobile
   > gomobile init

1. Place github.com/hyangah/mgodoc/godoc under GOPATH/src.

   > go get -d -u github.com/hyangah/mgodoc/godoc

2. Open the Android Studio project in github.com/hyangah/mgodoc/android.

3. Open the godoc/build.gradle file, and fill in the GOPATH and GO fields.

4. Build and run.


Godoc package

This directory contains a go package that works as a thin wrapper of the godoc package.
(https://golang.org/x/tools/godoc) It also includes an asset file, assets/go.zip, which
is an archive of GOROOT files the golang.org/x/tools/godoc package requires.

  * godoc.go:

  this file has code to load the asset file as a zip FS, initializes the godoc
  Corpus, and enables necessary http handlers implemented in golang.org/x/tools/godoc.
  Its function is similar to the https://golang.org/x/tools/cmd/godoc main package.

  Due to limited resources in mobile platform, CPU/memory intensive features
  such as indexing (so searching) and static analysis are not enabled yet.

  * http.go:

  this file includes the exported 'Serve' function, which is called by
  the WebView client in Java or Objective-C code. The current implementation
  uses []byte to return the web page contents. Once the gomobile bind
  starts to support more advanced types, I plan to switch to io.Reader
  or IOStream-like interface.

  This file shows how easy it is to convert an existing http web server
  into a local web data store. The powerful and flexible HTTP package of Go
  allows to use a custom transport and communicate with the web server
  without going through the full network stack.

  * zipfs.go:

  this is a slightly modified copy of golang.org/x/tools/godoc/vfs/zipfs
  that can take zip.Reader instead of zip.ReadCloser.

  The golang.org/x/tools/godoc package supports various virtual file systems.
  The zipfs implements the file system interface on top of a zip file.
  The godoc package requires go package files and resources to be located
  in GOROOT, so this mobile godoc app archives the necessary files into
  a go.zip file and packages it as an asset file. Because assets files
  cannot be accessed through the regular file operations but through
  the golang.org/x/mobile/asset package, I couldn't use the zip package
  to construct a zip.ReadCloser. As a result, I had to vendor the zipfs
  and modify it to accept zip.Reader instead.

  * zip.bash:

  A bash script to produce the go.zip asset file from the current GOROOT.


Android directory

The /android directory hosts the Android Studio project.
It consists of two modules, 'app' and 'godoc'.

The godoc module is a .AAR library module whose artifact is a .AAR file. Unlike the
traditional .AAR library module Android Studio creates to serve a static .AAR file,
however, this module compiles github.com/hyangah/mgodoc/godoc package by invoking
the `gomobile bind` command during build, and recreate the .AAR file from the current
source.

The app module contains the Java source code for the main activity. It depends on
the godoc module (See the dependency section of app/build.gradle).
The MainActivity.MyWebViewClient class (in app/src/main/org/golang/example/godoc/MainActivity.java)
implements the logic to intercept the web request for the host 'golang.org' and
serve the contents returned from Go. The github.com/hyangah/mgodoc/godoc.Serve function
is mapped to go.godoc.Godoc.Serve function.


IOS directory

The /ios directory hosts the Xcode iOS project. (TODO: upload the source)
*/
package mgodoc
