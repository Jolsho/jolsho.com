import { useCallback, useEffect, useRef, useState } from 'react';

type Msg = {
    code: number;
    timestamp: string;
    text: string;
};

export type ChatProps = {
    setError: (e: string) => void;
    isLive: boolean,
    src: string,
};

type ChatHooks = {
    msgs: Msg[];
    text: string;
    setText: (t: string) => void;
    submit: (e: React.FormEvent) => void;
};

const useChat = (props: ChatProps): ChatHooks => {
    const [msgs, setMsgs] = useState<Msg[]>([]);
    const [text, setText] = useState('');
    const socketRef = useRef<WebSocket | null>(null);
    const reconnectAttempts = useRef(0);
    const maxReconnectAttempts = 5;
    const { setError, src, isLive } = props;

    const connect = useCallback(() => {
        if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) return;


        const socket = new WebSocket(src+"/chat?room=jolsho");
        socketRef.current = socket;

        socket.addEventListener('open', () => {
            reconnectAttempts.current = 0; // reset attempts on success
            setError('');
        });

        socket.addEventListener('message', (event) => {
            try {
                const msg: Msg = JSON.parse(event.data);
                setMsgs((prev) => [...prev, msg]);
            } catch {
                console.error('Invalid message format:', event.data);
            }
        });

        socket.addEventListener('error', () => {
            setError('Chat failed to connect... trying again');
        });

        socket.addEventListener('close', () => {
            if (reconnectAttempts.current < maxReconnectAttempts) {
                const timeout = Math.min(1000 * 2 ** reconnectAttempts.current, 10000); // exponential backoff up to 10s
                reconnectAttempts.current += 1;
                setError(`Reconnecting to chat... attempt ${reconnectAttempts.current}`);
                setTimeout(connect, timeout);
            } else {
                setError('Unable to reconnect. Please refresh.');
            }
        });
    }, [src, isLive]);

    useEffect(() => {
        if (isLive) connect();
        return () => socketRef.current?.close();
    }, [connect, isLive]);

    const submit = useCallback(
        (e: React.FormEvent) => {
            e.preventDefault();
            if (!text.trim() || !socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) return;

            const now = new Date();
            const msg = { 
                timestamp: `${now.getHours()}:${now.getMinutes()}`, 
                text,
            };
            socketRef.current.send(JSON.stringify(msg));
            setText('');
        },
        [text]
    );
    return { msgs, text, setText, submit };
};

export default useChat;

