export interface RawMessage {
    id: string;
    sender: string;
    chat: string;
    content: string;
    timestamp: string;
    filename: string | null;
}

export type RawMessages = RawMessage[];

export type Point = { x: number; y: number };
