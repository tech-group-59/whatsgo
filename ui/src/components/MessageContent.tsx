import {RawMessage} from "../types.ts";
import {useState} from "react";
import {HighlightText} from "./HighlightText.tsx";

export const MessageContent = ({lastContent, message, className}: {
    lastContent: string,
    message: RawMessage,
    className: string
}) => {
    const [isFormatted, setIsFormatted] = useState(false);
    return (
        <div className={className} style={{
            position: 'relative',
        }}>
            <div onClick={() => {
                setIsFormatted(!isFormatted);
            }} style={{
                position: 'absolute',
                left: -10,
                top: -12,
                fontWeight: 'bold',
                cursor: 'pointer',
                color: 'grey',
                userSelect: 'none',
            }}>{isFormatted ? 'U' : 'F'}</div>
            {!isFormatted ? <div>
                    <HighlightText text={message.content} highlight={lastContent}/>
                </div> :
                <pre>
                {message.content}
            </pre>}
        </div>
    );
}