#!/bin/sh

cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. 
echo "==> Running selective tests..."
go test ./internal/pkg/archive_helper/
go test ./internal/pkg/calculate_curve_correlation
# go test ./internal/pkg/config
# go test ./internal/pkg/debug_view
go test ./internal/pkg/decode
go test ./internal/pkg/dtw
echo "ERROR:  Emby server needed to run emby_api test, skipping.............."
# go test ./internal/pkg/emby_api # fail
# go test ./internal/pkg/ffmpeg_helper # fail
# go test ./internal/pkg/frechet 
# go test ./internal/pkg/global_value
go test ./internal/pkg/gss
# go test ./internal/pkg/hot_fix
go test ./internal/pkg/imdb_helper
# go test ./internal/pkg/language
# go test ./internal/pkg/log_helper
# go test ./internal/pkg/my_util # will produce Log dir
echo "ERROR:  Proxy needed for proxy_helper, skipping.............."
# go test ./internal/pkg/url_connectedness_helper # fail
# go test ./internal/pkg/random_useragent
# go test ./internal/pkg/regex_things
echo "ERROR:  Proxy needed for rod_helper,  skipping.............."
# go test ./internal/pkg/rod_helper # fail
# go test ./internal/pkg/sqlite
# go test ./internal/pkg/sub_formatter # fail
echo "ERROR:  srt and ass has differet len of dialogue sub_helper_test.go"
go test ./internal/pkg/sub_helper
go test ./internal/pkg/sub_parser_hub
# fixer_test needs better files and audio
# go test ./internal/pkg/sub_timeline_fixer # fail
go test ./internal/pkg/vad
# go test ./internal/pkg/vosk_api

# 1. emby_api_test: need emby
# 2. ffmpeg_helper: will generate test files
# 3. sub_timeline_change/sub_format_changer_test.go wrong other normal

echo "==> Done..."
