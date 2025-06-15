function fmtTime(s) {
    const d = new Date(s * 1000);
    if (s > 600) {
	return d.toISOString().slice(11, 19);
    }
    return d.toISOString().slice(14, 19);
}

var audio = document.querySelector("audio")

var titleEl;
var coverImgEl;
var trackEls;
var currentTrackIdx;

function setTrack(idx) {
    trackEls[currentTrackIdx].classList.remove("currentTrack");
    currentTrackIdx = idx;
    const trackEl = trackEls[idx];
    trackEls[currentTrackIdx].classList.add("currentTrack");
    audio.src = trackEl.dataset.url;
    audio.title = trackEl.dataset.title;
    titleEl.innerText = trackEl.dataset.title;
    
    if ('mediaSession' in navigator) {
	let meta = {
            title: trackEl.dataset.title,
            artist: "",
            album: ""
	};
	if (coverImgEl) {
            meta.artwork = [{ src: coverImgEl.src }]
	}
	navigator.mediaSession.metadata = new MediaMetadata(meta);
    }
    saveTrack()
}

function play() {
    audio.currentTime = lastElementTime();
    audio.play();
    trackEls[currentTrackIdx].classList.add("playing");
}

function pause() {
    audio.pause();
    trackEls[currentTrackIdx].classList.remove("playing");
}

function prev() {

    pause();
    setTrack(currentTrackIdx - 1);
    play();
}

function next() {

    pause();
    setTrack(currentTrackIdx + 1);
    play();
}

function lastTrack() {
    var idx = localStorage.getItem(window.location.pathname);
    return idx == null ? 0 : idx;
}

function saveTrack() {
    localStorage.setItem(window.location.pathname, currentTrackIdx);
}

function lastElementTime() {
    var time = localStorage.getItem(`${window.location.pathname}:${currentTrackIdx}`);
    return time == null ? 0 : time;
}

function saveElementTime() {
    localStorage.setItem(`${window.location.pathname}:${currentTrackIdx}`, audio.currentTime);
}



function initPlayer() {
    titleEl = document.querySelector(".title");
    audio = document.querySelector("audio")

    coverImgEl = document.querySelector(".cover > img");
    trackEls = document.querySelectorAll(".track");
    if (trackEls.length == 0) {
	return;
    }
    currentTrackIdx = 0;

    setTrack(lastTrack());

    let mouseDownOnSlider = false;

    audio.addEventListener("timeupdate", () => {
	if (mouseDownOnSlider || !audio.duration) {
	    return;
	}
	saveElementTime()
    });
    audio.addEventListener("ended", () => {
	pause();
	if (currentTrackIdx < trackEls.length - 1) {
	    setTrack(currentTrackIdx + 1);
	    play();
	}
    });
    audio.addEventListener("pause", () => {
	trackEls[currentTrackIdx].classList.remove("playing");
    });
    audio.addEventListener("play", () => {
	trackEls[currentTrackIdx].classList.add("playing");
	trackEls[currentTrackIdx].scrollIntoView({ behavior: 'smooth', block: 'start' });
    });


    if ('mediaSession' in navigator) {
	// mediaSession is flaky in Chrome https://bugs.chromium.org/p/chromium/issues/detail?id=1337536
	navigator.mediaSession.setActionHandler('previoustrack', prev);
	navigator.mediaSession.setActionHandler('nexttrack', next);
	navigator.mediaSession.setActionHandler('pause', pause);
	navigator.mediaSession.setActionHandler('play', play);
	navigator.mediaSession.setActionHandler('seekto', function (data) {
	    audio.currentTime = data.seekTime;
	});
    }

    trackEls.forEach(el => el.addEventListener("click", event => {
	const trackEl = event.currentTarget;
	const targetIdx = parseInt(trackEl.dataset.index, 10);
	if (targetIdx == currentTrackIdx) {
	    if (audio.paused) {
		audio.play();
	    } else {
		audio.pause();
	    }
	    return;
	}
	pause();
	setTrack(targetIdx);
	play();
    }));
    
    titleEl.addEventListener("click", event => {
	if (audio.paused) {
	    audio.play();
	} else {
	    audio.pause();
	}
    })
}

window.addEventListener("DOMContentLoaded", initPlayer);
