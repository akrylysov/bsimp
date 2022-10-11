function fmtTime(s) {
  const d = new Date(s * 1000);
  if (s > 600) {
    return d.toISOString().slice(11, 19);
  }
  return d.toISOString().slice(14, 19);
}

function initPlayer() {
  const titleEl = document.querySelector(".title");
  const buttonPlayPauseEl = document.querySelector(".button-playpause");
  const progressEl = document.querySelector("input[type='range']");
  const timeElapsedEl = document.querySelector(".time-elapsed");
  const timeTotalEl = document.querySelector(".time-total");
  const buttonPrevEl = document.querySelector(".button-prev");
  const buttonNextEl = document.querySelector(".button-next");
  const coverImgEl = document.querySelector(".cover > img");
  const trackEls = document.querySelectorAll(".track");
  if (trackEls.length == 0) {
    return;
  }
  var currentTrackIdx = 0;

  if (trackEls.length > 1) {
    buttonNextEl.classList.remove("disabled");
  }

  const audio = new Audio();

  function setTrack(idx) {
    currentTrackIdx = idx;
    const trackEl = trackEls[idx];
    audio.src = trackEl.dataset.url;
    titleEl.innerText = trackEl.dataset.title;

    if (idx == 0) {
      buttonPrevEl.classList.add("disabled");
    } else {
      buttonPrevEl.classList.remove("disabled");
    }

    if (idx == trackEls.length - 1) {
      buttonNextEl.classList.add("disabled");
    } else {
      buttonNextEl.classList.remove("disabled");
    }

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
  }

  function play() {
    audio.play();
    buttonPlayPauseEl.classList.add("playing");
    trackEls[currentTrackIdx].classList.add("playing");
  }

  function pause() {
    audio.pause();
    buttonPlayPauseEl.classList.remove("playing");
    trackEls[currentTrackIdx].classList.remove("playing");
  }

  setTrack(0);

  let mouseDownOnSlider = false;

  audio.addEventListener("loadeddata", () => {
    progressEl.value = 0;
  });
  audio.addEventListener("timeupdate", () => {
    if (mouseDownOnSlider || !audio.duration) {
      return;
    }
    progressEl.value = audio.currentTime / audio.duration * 100;
    timeElapsedEl.textContent = fmtTime(audio.currentTime);
    timeTotalEl.textContent = fmtTime(audio.duration);
  });
  audio.addEventListener("ended", () => {
    pause();
    if (currentTrackIdx < trackEls.length - 1) {
      setTrack(currentTrackIdx + 1);
      play();
    }
  });
  audio.addEventListener("pause", () => {
    buttonPlayPauseEl.classList.remove("playing");
    trackEls[currentTrackIdx].classList.remove("playing");
  });
  audio.addEventListener("play", () => {
    buttonPlayPauseEl.classList.add("playing");
    trackEls[currentTrackIdx].classList.add("playing");
  });

  buttonPlayPauseEl.addEventListener("click", () => {
    if (audio.paused) {
      play();
    } else {
      pause();
    }
  });

  progressEl.addEventListener("change", () => {
    const pct = progressEl.value / 100;
    audio.currentTime = (audio.duration || 0) * pct;
  });
  progressEl.addEventListener("mousedown", () => {
    mouseDownOnSlider = true;
  });
  progressEl.addEventListener("mouseup", () => {
    mouseDownOnSlider = false;
  });


  function prev() {
    if (buttonPrevEl.classList.contains("disabled")) {
      return;
    }
    pause();
    setTrack(currentTrackIdx - 1);
    play();
  }

  function next() {
    if (buttonNextEl.classList.contains("disabled")) {
      return;
    }
    pause();
    setTrack(currentTrackIdx + 1);
    play();
  }

  buttonPrevEl.addEventListener("click", prev);
  buttonNextEl.addEventListener("click", next);

  if ('mediaSession' in navigator) {
    // mediaSession is flaky in Chrome https://bugs.chromium.org/p/chromium/issues/detail?id=1337536
    navigator.mediaSession.setActionHandler('previoustrack', prev);
    navigator.mediaSession.setActionHandler('nexttrack', next);
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
}

window.addEventListener("DOMContentLoaded", initPlayer);
