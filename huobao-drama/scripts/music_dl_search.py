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
from music_dl.source import MusicSource
from music_dl.song import BasicSong


def _disable_downloads():
    BasicSong._download_file = lambda *args, **kwargs: None
    BasicSong.download_song = lambda self: None
    BasicSong.download_lyrics = lambda self: None
    BasicSong.download_cover = lambda self: None


def _resolve_url(song):
    if getattr(song, "song_url", ""):
        return song.song_url
    try:
        song.download()
    except Exception:
        return ""
    return getattr(song, "song_url", "") or ""


def main():
    if len(sys.argv) < 5:
        print(json.dumps({"error": "args"}))
        return
    keyword = sys.argv[1]
    page = int(sys.argv[2])
    page_size = int(sys.argv[3])
    sources = [s for s in sys.argv[4].split(",") if s]
    resolve_limit = int(sys.argv[5]) if len(sys.argv) > 5 else 0

    # music-dl 库会直接向 stdout/stderr 打印日志，需屏蔽以保证只输出 JSON。
    with contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
        config.init()
        config.set("keyword", keyword)
        config.set("number", page * page_size)
        config.set("nomerge", False)

        _disable_downloads()

        ms = MusicSource()
        songs = ms.search(keyword, sources)

    start = (page - 1) * page_size
    end = start + page_size
    page_songs = songs[start:end]

    items = []
    for idx, s in enumerate(page_songs):
        url = ""
        if resolve_limit > 0 and idx < resolve_limit:
            with contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
                url = _resolve_url(s)
        item = {
            "id": str(getattr(s, "id", "") or ""),
            "title": getattr(s, "title", ""),
            "artist": getattr(s, "singer", ""),
            "album": getattr(s, "album", ""),
            "duration": getattr(s, "duration", ""),
            "source": getattr(s, "source", ""),
            "song_url": url,
            "mid": str(getattr(s, "mid", "") or ""),
            "hash": str(getattr(s, "hash", "") or ""),
            "content_id": str(getattr(s, "content_id", "") or ""),
        }
        items.append(item)

    print(json.dumps({"items": items, "total": len(songs)}))


if __name__ == "__main__":
    main()
