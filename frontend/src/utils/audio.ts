import { AUDIO_ASSETS } from '../config/constants';

// Audio Files Configuration (Loaded from /public/assets/audio)
const alarmAudio = new Audio(AUDIO_ASSETS.alarm);
const gentleChimeAudio = new Audio(AUDIO_ASSETS.gentleChime);
const happyChimeAudio = new Audio(AUDIO_ASSETS.happyChime);

// Preload audio files to avoid latency
alarmAudio.preload = 'auto';
gentleChimeAudio.preload = 'auto';
happyChimeAudio.preload = 'auto';

// Allow concurrent playback by cloning the audio node if needed
function playSound(audio: HTMLAudioElement) {
    if (!audio) return;

    // Reset time to start to allow rapid re-triggers
    audio.currentTime = 0;

    const playPromise = audio.play();
    if (playPromise !== undefined) {
        playPromise.catch((error) => {
            console.warn(`[Audio] Failed to play sound: ${error.message} - Please ensure files exist in public/assets/audio/`);
        });
    }
}

export function playAlarmSound() {
    playSound(alarmAudio);
}

export function playGentleChime() {
    playSound(gentleChimeAudio);
}

export function playHappyChime() {
    playSound(happyChimeAudio);
}