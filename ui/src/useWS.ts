import {useState, useEffect, useRef} from 'react';

type UseWSProps = {
    url: string;
    onMessage?: (event: MessageEvent) => void;
    onOpen?: () => void;
    onClose?: () => void;
};

const noop = () => {
};

const useWS = ({
                   url, onMessage = () => {
    }, onOpen = noop, onClose = noop
               }: UseWSProps) => {
    const [isConnected, setIsConnected] = useState(false);
    const [lastMessage, setLastMessage] = useState<string>('');
    const ws = useRef<WebSocket | null>(null);
    const reconnectTimeout = useRef<number | null>(null);

    const disconnectWS = () => {
        ws.current?.close();
        setIsConnected(false);
        ws.current = null;
    };

    const reconnect = () => {
        if (reconnectTimeout.current) {
            clearInterval(reconnectTimeout.current);
        }
        reconnectTimeout.current = setInterval(() => {
            if (ws.current === null || ws.current?.readyState === WebSocket.CLOSED) {
                console.log('reconnecting...');
                connectWS();
            }
        }, 2000);
    }

    const connectWS = () => {
        if (isConnected) return;

        console.log(`[${new Date().toLocaleTimeString()}] WS creating`);

        ws.current = new WebSocket(url);

        ws.current.onopen = () => {
            setIsConnected(true);
            if (onOpen) {
                onOpen();
            }
        };

        ws.current.onclose = () => {
            if (onClose) {
                onClose();
            }
            ws.current = null;
        };

        ws.current.onmessage = (event: MessageEvent) => {
            if (onMessage) {
                onMessage(event);
            }
            setLastMessage(event.data);
        };

        ws.current.onerror = (e) => {
            console.error('ws error:', e);
            disconnectWS();
        };
        reconnect();
    };

    useEffect(() => {

            return () => {
                if (reconnectTimeout.current) {
                    clearTimeout(reconnectTimeout.current);
                }
                disconnectWS();
            }
        },
        []
    );

    return {
        isConnected,
        connectWS,
        disconnectWS,
        lastMessage,
    };
};

export default useWS;
