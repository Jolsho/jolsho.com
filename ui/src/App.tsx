import { useEffect, useState } from 'react';
import './App.css';
import Video from './Live';

export type Card = {
    id: number;
    image: string;
    heading: string;
    description: string;
    finePrint: string;
};
export type Stream = {
    Name: string,
    IsLive: boolean,
    Title: string,
    Viewers: number,
};

function App() {
    const src = 'https://localhost:443';

    const [_page, setPage] = useState('main');
    const [stream, setStream] = useState<Stream|null>(null);

    useEffect(() => {
        let intervalId;

        async function handleIsLive() {
            try {
                const res = await fetch(src + '/isLive?room=jolsho');
                if (!res.ok) throw new Error('Network response was not ok');
                const data:Stream = await res.json();
                setStream((prev) => {
                    if (data.IsLive !== prev?.IsLive) {
                        return data
                    } else if (data.Title !== prev?.Title) {
                        return data
                    }
                    return prev
                });
            } catch (err) {
                console.error('Failed to check live status:', err);
            }
        }

        // Initial check
        handleIsLive();

        // Poll every 10 seconds
        intervalId = setInterval(handleIsLive, 10000);

        // Cleanup on unmount
        return () => clearInterval(intervalId);
    }, [src]);

    return (
        <div className="menu pale">
            <div className="socials-links">
                <a
                    href="https://github.com/jolsho"
                    target="_blank"
                    className="social-name"
                >
                    Github
                </a>
                <a
                    href="https://www.linkedin.com/in/joshua-o-671837381"
                    target="_blank"
                    className="social-name"
                >
                    LinkedIn
                </a>
                <a
                    href="https://open.spotify.com/user/jolsho"
                    target="_blank"
                    className="social-name"
                >
                    Spotify
                </a>
            </div>

            <p className="font heading" onClick={() => setPage('about')}>
                JOSHUA
            </p>
            {/* <p className="font sub-heading"> */}
            {/*     Software Engineer | Cryptography{' '} */}
            {/* </p> */}
            <div className="body-row-split">
                <div className="video-container pale">
                    <div className="video-header">
                        <div className={`is-live-light ${stream?.IsLive && 'live'}`} />
                        <div className="video-description">
                            <p> {stream?.Title ?? "Not Currently Live"} </p>
                        </div>
                    </div>
                    <Video src={src} stream={stream} />
                </div>
            </div>
        </div>
    );
}

export default App;
