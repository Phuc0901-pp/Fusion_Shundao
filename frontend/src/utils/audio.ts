// Audio & TTS Utilities

// ─── Tone Generation (Web Audio API) ───────────────────────────────
function playTone(frequency: number, duration: number, volume = 0.2, type: OscillatorType = 'sine') {
    try {
        const AudioContext = window.AudioContext || (window as any).webkitAudioContext;
        const ctx = new AudioContext();
        const osc = ctx.createOscillator();
        const gain = ctx.createGain();

        osc.connect(gain);
        gain.connect(ctx.destination);

        osc.frequency.value = frequency;
        osc.type = type;

        gain.gain.setValueAtTime(volume, ctx.currentTime);
        gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + duration);

        osc.start(ctx.currentTime);
        osc.stop(ctx.currentTime + duration);
    } catch { /* Audio not supported */ }
}

export function playAlarmSound() {
    playTone(880, 0.3, 0.3, 'square');
    setTimeout(() => playTone(880, 0.3, 0.3, 'square'), 350);
}

export function playGentleChime() {
    playTone(523, 0.4, 0.15, 'sine'); // C5
    setTimeout(() => playTone(392, 0.5, 0.12, 'sine'), 300); // G4
}

export function playHappyChime() {
    playTone(523, 0.3, 0.15, 'sine');  // C5
    setTimeout(() => playTone(659, 0.3, 0.15, 'sine'), 250); // E5
    setTimeout(() => playTone(784, 0.4, 0.18, 'sine'), 500); // G5
}

// ─── TTS Engine ────────────────────────────────────────────────────
let vietnameseVoice: SpeechSynthesisVoice | null = null;

function loadVoices() {
    if (typeof window === 'undefined' || !window.speechSynthesis) return;
    const voices = window.speechSynthesis.getVoices();
    // Prioritize "Google Tiếng Việt" or "Microsoft HoaiMy" (Edge)
    vietnameseVoice = voices.find(v => v.lang === 'vi-VN' && (v.name.includes('Google') || v.name.includes('Microsoft')))
        || voices.find(v => v.lang === 'vi-VN')
        || null;
}

if (typeof window !== 'undefined' && 'speechSynthesis' in window) {
    window.speechSynthesis.onvoiceschanged = loadVoices;
    loadVoices();
}

// Fallback: Google Translate TTS (Online) - Robust Multi-domain
function playOnlineTTS(text: string) {
    try {
        const encodedText = encodeURIComponent(text);

        // 1. Try .vn (Low latency for Vietnam)
        const urlVN = `https://translate.google.com.vn/translate_tts?ie=UTF-8&client=tw-ob&tl=vi&q=${encodedText}`;
        const audio = new Audio(urlVN);

        audio.play().catch(() => {
            // 2. Try .com (Standard)
            const urlCOM = `https://translate.google.com/translate_tts?ie=UTF-8&client=tw-ob&tl=vi&q=${encodedText}`;
            new Audio(urlCOM).play().catch(() => {
                // 3. Try googleapis (Alternative ID)
                const urlAPI = `https://translate.googleapis.com/translate_tts?client=gtx&ie=UTF-8&tl=vi&q=${encodedText}`;
                new Audio(urlAPI).play().catch(e => console.error('All TTS sources failed:', e));
            });
        });
    } catch { /* Ignore errors */ }
}

export function speakAlert(message: string) {
    try {
        if ('speechSynthesis' in window) {
            window.speechSynthesis.cancel();

            if (vietnameseVoice) {
                const utterance = new SpeechSynthesisUtterance(message);
                utterance.lang = 'vi-VN';
                utterance.voice = vietnameseVoice;
                utterance.rate = 1.0;
                utterance.volume = 1.0;
                utterance.pitch = 1.0;
                window.speechSynthesis.speak(utterance);
            } else {
                playOnlineTTS(message);
            }
        } else {
            playOnlineTTS(message);
        }
    } catch {
        playOnlineTTS(message);
    }
}
