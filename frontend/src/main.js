import './style.css';
import './app.css';

import {RTSP, CloseRTSP} from '../wailsjs/go/main/App';
import {EventsOn, EventsEmit} from '../wailsjs/runtime/runtime';

let mediaSource, sourceBuffer;

function close() {
    mediaSource = null;
    sourceBuffer = null;
    CloseRTSP();
}

window.OnPlayGo = function () {
    close();
    let username = document.getElementById("username").value;
    let password = document.getElementById("pw").value;
    let url = document.getElementById("url").value;

    if(username && password) {
        url = url.replace("rtsp://", `rtsp://${username}:${password}@`);
    }
    RTSP(url);
};

EventsOn("OnInit", function(meta, init) {
    const video = document.getElementById("remoteVideo");
    mediaSource = new MediaSource();
    mediaSource.addEventListener("sourceopen", function() {
        if(mediaSource.sourceBuffers.length > 0) {
            return;
        }

        sourceBuffer = mediaSource.addSourceBuffer(`video/mp4; codecs="${meta}"`);
        let updateendCallback = function() {
            sourceBuffer.removeEventListener("updateend", updateendCallback);
            EventsEmit("OnUpdateEnd");
        };
        sourceBuffer.mode = "segments"
        sourceBuffer.addEventListener("updateend", updateendCallback);
        pushBuffer(init);
        video.play();
    });
    video.src = window.URL.createObjectURL(mediaSource);
});

EventsOn("OnFrame", function(frame) {
    pushBuffer(frame);
});

function pushBuffer(encoded) {
    if(sourceBuffer) {
        const decoded = atob(encoded);
        const arr = new Uint8Array(decoded.length);
        for(let i = 0; i < decoded.length; i++) {
            arr[i] = decoded.charCodeAt(i);
        }

        sourceBuffer.appendBuffer(arr.buffer);
    }
}
