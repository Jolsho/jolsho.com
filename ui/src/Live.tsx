import { useRef, useEffect, useState } from 'react';
import './live.css';
import useChat, { type ChatProps } from './Chat';
import Hls from 'hls.js';
import type { Stream } from './App';

export type VideoProps = {
    stream: Stream|null;
    src: string;
};

function Video({ src, stream }: VideoProps) {
    const videoRef = useRef<HTMLVideoElement | null>(null);
    const [error, setError] = useState(false);

    useEffect(() => {
        const video = videoRef.current;
        if (!video) return;

        const videoPath = `${src}/hls/jolsho/index.m3u8`;
        setError(false); // reset error on src/isLive change

        const handleError = () => {
            console.warn('No live stream or VOD available currently.');
            setError(true);
        };

        if (video.canPlayType('application/vnd.apple.mpegurl')) {
            // Safari/native HLS
            video.src = videoPath;
            video.play().catch(() => handleError());
            video.onerror = handleError;
            return;
        }

        if (Hls.isSupported()) {
            const hls = new Hls();
            hls.loadSource(videoPath);
            hls.attachMedia(video);

            const timeout = setTimeout(() => {
                console.warn('HLS manifest load timed out');
                hls.destroy();
                handleError();
            }, 10000);

            hls.on(Hls.Events.MANIFEST_PARSED, () => {
                clearTimeout(timeout);
                video.addEventListener('canplay', () => {
                    video.play().catch(() => {
                        console.error("VIDEO PLAY");
                        handleError();
                    });
                });
            });

            hls.on(Hls.Events.ERROR, (_event, data) => {
                if (data.fatal) {
                    clearTimeout(timeout);
                    hls.destroy();
                    console.error("fatal:", data.type, data.details,data.response);
                    handleError();
                }
            });

            return () => {
                clearTimeout(timeout);
                hls.destroy();
            };
        } else {
            console.error("ELSED");
            // HLS not supported and native play fails
            handleError();
        }
    }, [src, stream?.IsLive]);

    return (
        <div className="live-container">
            {error ? (
                <>
                {/* <img src='./screen.jpg' className='live-video'> */}
                <div className="live-video backdrop font">
                    No live stream or VOD available currently.
                </div>
                {/* </img> */}
                </>
            ) : (
                <video
                    ref={videoRef}
                    className="live-video"
                    controls
                    muted
                />
            )}
            <LiveChat src={src} stream={stream} />
        </div>
    );
};
export default Video;

function LiveChat ({ stream, src }: VideoProps) {
    const [error, setError] = useState('');
    const chatProps: ChatProps = { 
        setError, src, 
        isLive: stream?.IsLive ?? false
    };
    const { msgs, text, setText, submit } = useChat(chatProps);

    const scrollableChat = useRef<HTMLDivElement | null>(null);
    const [isOpen, setIsHidden] = useState(false);

    useEffect(() => {
        const chat = scrollableChat.current;
        if (chat) {
            chat.scrollTop = chat.scrollHeight;
        }
    }, [msgs]);

    return (
        <>
        <div className='chat-opener' onClick={()=>setIsHidden(!isOpen)}>
            <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                className={`arrow ${isOpen ? 'rotated' : ''}`}
            >
                <polyline points="6 9 12 15 18 9"></polyline>
            </svg>
        </div>
        <div className={`live-chat ${isOpen ? 'open' : ''}`}>
            <div
                className={`chat-messages ${isOpen ? 'collapsed' : ''}`}
                ref={scrollableChat}
            >
                {msgs.map((m, _) => {
                    return (
                        <div className="msg">
                            <h1> {m.timestamp} </h1>
                            <p> {m.text} </p>
                        </div>
                    );
                })}
            </div>
            <input
                type="text"
                className="chat-input pale"
                placeholder="Type a message..."
                value={text}
                onKeyDown={(e) => {
                    if (e.key == 'Enter') {
                        submit(e);
                    }
                }}
                onChange={(e) => setText(e.target.value)}
            />
            <div className="chat-error">
                <p>{error}</p>
            </div>
        </div>
        </>
    );
};
