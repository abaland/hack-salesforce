# salesforce-chronus sync

![testing](https://github.com/abaland/hack-salesforce/actions/workflows/testing.yml/badge.svg)

## Description
Parses timesheet from salesforce's TeamSpirit application and records the parsed data into the TIS chronus website.

SalesforceのTeamSpritiに登録されているデータを読み込み、TISのChronusツールに登録する。

## How it works

Since neither Salesforce nor Chronus have an available REST API or are using 
template rendering to show the data, the code uses ChromeDriver to log in as 
a human would do (but faster) into both website and parse/register the data 
through a mix of HTML parsing and JS.

SalesForceもChronusもREST APIとテンプレートレンダリング機能がついていないので、ChromeDriver
を使って人間のようにログインしてHTMLとJSを使って読み込み・書き込みをします。

## How to use
Once the tool is built, rename the template_config.ini file as config.ini, fill in
your credentials and just run the code.

プログラムをビルドしてからtemplate_config.iniをconfig.iniとしてリネームし、
ファイルの中にある認証情報を入力し.exeファイルを実行します。

## Build

### Install Golang
http://golang.jp/install

### Install ChromeDriver
Install [ChromeDriver](https://sites.google.com/a/chromium.org/chromedriver/downloads) and set Path.

### Build
```bash
git clone https://github.com/abaland/hack-salesforce 
cd hack-salesforce
go get -t ./src
go build -o chronus_sync.exe -v ./src
******edit config_file******
```

## How to use
```bash
./chronus_sync.exe
```
If config.ini is located in the same folder.

```bash
./chronus_sync.exe --config_path {config_path} 
```
otherwise.


## Notes/Missing features

The following features are still missing:
* 午前休・午後休： I almost never take those so I didn't have sample data to process them
* More than 3 breaks per day. Chronus just supports 3 aside from lunch breaks so I didn't add support for more.

下記の機能はまだ含まれていない：
* 午前休・午後休：あまり取らないのでどう登録されているかわからなくて開発できなかった。
* 休憩回数4回以上：Chronusでは休憩回数は最大3回（昼間以外）なので対応していない。

Please note the following:
* Break times too complicated might lead to bugs. Because Chronus has an invisible unchangeable break during 12:00-13:00, 
  i'm moving the salesforce lunch break to this interval, but there might be overlaps and so on.
  
注意点：
* 複雑すぎる休憩はバグになるかもしれない。Chronusでは12:00-13:00の昼休憩（非表示で変更不可）があるので、Salesforce
  の昼休憩の時間(デフォルト：13:00-14:00)を移動しているが、オーバーラップがないか確認していない。

# Refs
https://github.com/shotasym/hack-salesforce

This repository is forked from the above and served as a base. The goal of that
repo was to put info from a JSON file into Salesforce, so the process has mostly 
been inverted abd extended, but a lot of the code could still be reused.