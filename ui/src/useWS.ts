import { useState, useEffect, useRef } from 'react';

type UseWSProps = {
    url: string;
    onMessage?: (event: MessageEvent) => void;
    onOpen?: () => void;
    onClose?: () => void;
};

const noop = () => {};

const useWS = ({
    url,
    onMessage = () => {},
    onOpen = noop,
    onClose = noop,
}: UseWSProps) => {
    const [isConnected, setIsConnected] = useState(false);
    const [lastMessage, setLastMessage] = useState<string>('');
    const ws = useRef<WebSocket | null>(null);

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
            setIsConnected(false);
            if (onClose) {
                onClose();
            }
        };

        ws.current.onmessage = (event: MessageEvent) => {
            setLastMessage(event.data);
            if (onMessage) {
                onMessage(event);
            }
        };

        ws.current.onerror = (e) => {
            console.log(e);
            setIsConnected(false);
        };
    };

    const disconnectWS = () => {
        setIsConnected(false);
        ws.current?.close();
    };

    useEffect(
        () => () => {
            disconnectWS();
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
