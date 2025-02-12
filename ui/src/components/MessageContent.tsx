import {RawMessage} from "../types.ts";
import {useState} from "react";
import {HighlightText} from "./HighlightText.tsx";

export const MessageContent = ({lastContent, message, styles}: {
    lastContent: string,
    message: RawMessage,
    styles: (string | null)[],
}) => {
    const [isFormatted, setIsFormatted] = useState(false);
    const [className, color] = styles;
    return (
        <div className={className ?? undefined} style={{
            position: 'relative',
            backgroundColor: color ?? undefined,
        }}>
            {!!message.content && <div onClick={() => {
                setIsFormatted(!isFormatted);
            }} style={{
                position: 'absolute',
                left: -10,
                top: -12,
                fontWeight: 'bold',
                cursor: 'pointer',
                color: 'grey',
                userSelect: 'none',
            }}>{isFormatted ? 'U' : 'F'}</div>}
            {!isFormatted ? <div>
                    <HighlightText text={message.content} highlight={lastContent}/>
                </div> :
                <pre>
                {message.content}
            </pre>}
        </div>
    );
}