export interface RawMessage {
    id: string;
    sender: string;
    chat: string;
    content: string;
    timestamp: string;
    parsed_content: string;
}

export type RawMessages = RawMessage[];
