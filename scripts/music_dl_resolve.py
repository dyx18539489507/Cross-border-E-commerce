#!/usr/bin/env python3
import json
import io
import os
import sys
import contextlib

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
THIRD_PARTY = os.path.abspath(os.path.join(BASE_DIR, "..", "third_party"))
sys.path.insert(0, THIRD_PARTY)

from music_dl import config
from music_dl.song import BasicSong


def _disable_downloads():
    BasicSong._download_file = lambda *args, **kwargs: None
    BasicSong.download_song = lambda self: None
    BasicSong.download_lyrics = lambda self: None
    BasicSong.download_cover = lambda self: None


def main():
    if len(sys.argv) < 3:
        print(json.dumps({"error": "args"}))
        return
    source = sys.argv[1]
    payload = json.loads(sys.argv[2])

    # music-dl 库会直接向 stdout/stderr 打印日志，需屏蔽以保证只输出 JSON。
    with contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
        config.init()
        _disable_downloads()

    # 依赖 config.init() 初始化后再导入，避免第三方库在导入期读取未初始化配置
    from music_dl.addons.qq import QQSong
    from music_dl.addons.kugou import KugouSong
    from music_dl.addons.migu import MiguSong
    from music_dl.addons.netease import NeteaseSong

    song = None
    if source == "qq":
        song = QQSong()
        song.mid = payload.get("mid", "")
        song.id = payload.get("id", "")
        song.title = payload.get("title", "")
        song.singer = payload.get("artist", "")
    elif source == "kugou":
        song = KugouSong()
        song.hash = payload.get("hash", "")
        song.id = payload.get("id", "")
        song.title = payload.get("title", "")
        song.singer = payload.get("artist", "")
    elif source == "migu":
        song = MiguSong()
        song.content_id = payload.get("content_id", "")
        song.id = payload.get("id", "")
        song.title = payload.get("title", "")
        song.singer = payload.get("artist", "")
    elif source == "netease":
        song = NeteaseSong()
        song.id = payload.get("id", "")
        song.title = payload.get("title", "")
        song.singer = payload.get("artist", "")
    else:
        print(json.dumps({"error": "unknown source"}))
        return

    with contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
        try:
            song.download()
        except Exception:
            pass

    print(json.dumps({"url": getattr(song, "song_url", "") or ""}))


if __name__ == "__main__":
    main()
