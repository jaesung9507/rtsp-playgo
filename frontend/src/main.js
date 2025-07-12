import './style.css';
import './app.css';

import {RTSP, CloseRTSP} from '../wailsjs/go/main/App';
import {EventsOn, EventsEmit} from '../wailsjs/runtime/runtime';

let mediaSource, sourceBuffer;
let frameQueue = [];
let isAppending = false;

function close() {
    frameQueue = [];
    isAppending = false;
    mediaSource = null;
    sourceBuffer = null;
    CloseRTSP();
}

window.OnPlayGo = function () {
    close();
    let username = document.getElementById("username").value;
    let password = document.getElementById("pw").value;
    let url = document.getElementById("url").value;

    if (username && password) {
        url = url.replace("rtsp://", `rtsp://${username}:${password}@`);
    }
    RTSP(url);
};

EventsOn("OnInit", function (meta, init) {
    const video = document.getElementById("remoteVideo");
    mediaSource = new MediaSource();

    mediaSource.addEventListener("sourceopen", function () {
        if (mediaSource.sourceBuffers.length > 0) {
            return;
        }

        try {
            sourceBuffer = mediaSource.addSourceBuffer(`video/mp4; codecs="${meta}"`);
        } catch (e) {
            console.error("failed to add source buffer:", e);
            return;
        }
        
        sourceBuffer.mode = "segments";
        sourceBuffer.addEventListener("updateend", onUpdateEnd);
        sourceBuffer.addEventListener("error", (e) => console.error("source buffer error:", e));

        let initAppendDone = function() {
            sourceBuffer.removeEventListener("updateend", initAppendDone);
            EventsEmit("OnUpdateEnd");
            video.play().catch(e => console.error("failed to play video:", e));
        };
        sourceBuffer.addEventListener("updateend", initAppendDone);
        pushBuffer(init);
    });

    video.src = window.URL.createObjectURL(mediaSource);
    video.addEventListener("error", (e) => console.error("video error:", e));
});

EventsOn("OnFrame", function (frame) {
    pushBuffer(frame);
});

function pushBuffer(encoded) {
    const decoded = atob(encoded);
    const arr = new Uint8Array(decoded.length);
    for (let i = 0; i < decoded.length; i++) {
        arr[i] = decoded.charCodeAt(i);
    }
    frameQueue.push(arr.buffer);
    appendNextFrame();
}

function appendNextFrame() {
    if (isAppending || frameQueue.length === 0 || !sourceBuffer || sourceBuffer.updating) {
        return;
    }

    isAppending = true;
    const frame = frameQueue.shift();
    
    appendFrameData(frame);
}

function appendFrameData(frame) {
    try {
        sourceBuffer.appendBuffer(frame);
    } catch (e) {
        console.error("failed to append frame:", e);
        isAppending = false;
    }
}

function onUpdateEnd() {
    isAppending = false;
    appendNextFrame();
}