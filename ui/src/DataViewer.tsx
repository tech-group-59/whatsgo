import {createUseStyles} from 'react-jss';
import {useEffect, useMemo, useRef, useState} from "react";

import moment from 'moment';
import useWS from "./useWS.ts";
import {Box, Button, Modal} from "@mui/material";
import {RawMessage, RawMessages} from "./types";
import {getYesterdaysDate, parseCoordinatesFromContent, parseDateTime} from "./helpers";
import {MessageContent} from "./components/MessageContent";
import {ParsedContent} from "./components/ParsedContent.tsx";


let _host = ``;

const {env} = import.meta;

if (env && env.VITE_HOST) {
    _host = env.VITE_HOST;
} else {
    const urlParams = new URLSearchParams(window.location.search);
    const hostFromUrl = urlParams.get('host');
    if (hostFromUrl) {
        _host = hostFromUrl;
    }
}
const host = _host;


const useStyles = createUseStyles({
    wrap: {
        display: 'flex',
        flexDirection: 'column',
        padding: '1rem',
    },
    form: {},
    inputGroupRow: {
        padding: '0 0 0.5rem 0',
        display: 'flex',
        flexDirection: 'row',
        alignItems: 'flex-end',
        gap: '0.5rem',
    },
    inputGroup: {
        display: 'flex',
        flexDirection: 'column',
    },
    spinnerWrap: {
        padding: '1rem',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
    },
    spinner: {
        width: '100px',
        height: '100px',
        border: '5px solid #f3f3f3',
        borderRadius: '50%',
        borderTop: '5px solid #3498db',
        animation: 'spin 2s linear infinite',
    },
    dataTable: {
        padding: '0.5rem',
        border: '1px solid #f3f3f3',
        borderRadius: '5px',
    },
    table: {
        borderSpacing: 0,
    },
    dataRow: {
        '&:hover': {
            background: 'rgba(243,243,243,0.2)',
        },
        '&:last-child td': {
            'border-bottom': 'none',
        },
    },
    td: {
        padding: '0.5rem',
        'border-bottom': '1px solid #f3f3f3',
    },
    selected: {
        background: 'rgba(114,0,0,0.74)',
    }
})

const notificationSound = '/notify.mp3';


function DataViewer() {
    const classes = useStyles();

    const [open, setOpen] = useState(false);

    const [dateFrom, setDateFrom] = useState(getYesterdaysDate());
    const [dateTo, setDateTo] = useState(new Date());
    const [content, setContent] = useState('');
    const [chats, setChats] = useState<{ [key: string]: string }>({});
    const [selectedChat, setSelectedChat] = useState('');
    const [messages, setMessages] = useState<RawMessages>([]);
    const [lastMessageTs, setLastMessageTs] = useState('');
    const [loading, setLoading] = useState(false);
    const [justOpened, setJustOpened] = useState(true);
    const [lastContent, setLastContent] = useState('');
    const [audioReady, setAudioReady] = useState(false);
    const [pullNewMessages, setPullNewMessages] = useState(true);
    const [contentGroups, setContentGroups] = useState<string[]>([]);
    const [selectedContentGroups, setSelectedContentGroups] = useState<string[]>([]);
    const messageRef = useRef<string | null>(null);
    const messagesRef = useRef<Set<string>>(new Set());
    const [selectedMessageIds, setSelectedMessageIds] = useState<string[]>([]);

    const {connectWS, disconnectWS, lastMessage} = useWS({
        url: `ws://${host}/ws`,
        onOpen: () => {
            console.log('connected to device');
        },
        onClose: () => {
            console.log('Disconnected from device');
        },
        onMessage: (event) => {
            messageRef.current = event.data;
        }
    });


    useEffect(() => {
        if (messageRef.current) {
            if (justOpened || !pullNewMessages) {
                return;
            }
            const msg = JSON.parse(messageRef.current) as RawMessage;

            if (messagesRef.current.has(msg.id)) {
                return;
            }
            messagesRef.current.add(msg.id);

            // check if lastContent is in the message
            if (msg.content.toLowerCase().includes(lastContent.toLowerCase()) || !lastContent) {
                // prepend the new message to the list
                setMessages([msg, ...messages]);

                if (audioReady) {
                    const audio = new Audio(notificationSound);
                    audio.play().catch((error) => console.error("Failed to play the sound:", error));
                }

                new Notification('New message', {
                    body: msg.content,
                });
            }
            messageRef.current = null;
        }
    }, [lastMessage]);

    useEffect(() => {
        if (messages.length) {
            const groups = messages.reduce((acc: string[], message) => {
                if (!message.content) {
                    return acc;
                }
                const firstLine = (message.content.split('\n')[0]).trim();
                if (!acc.includes(firstLine)) {
                    acc.push(firstLine);
                }
                return acc;
            }, []);
            setContentGroups(groups);
        }

    }, [messages]);

    useEffect(() => {
        // clear selectedMessageIds when selectedChat changes
        setSelectedMessageIds([]);
    }, [selectedChat]);

    const handleUserInteraction = async () => {
        const audio = new Audio(notificationSound);
        try {
            audio.volume = 0;
            await audio.play();
            setAudioReady(true);
        } catch (error) {
            console.error("Failed to play the sound:", error);
        }
    };

    useEffect(() => {
        // Function to handle requesting notification permission
        const requestNotificationPermission = async () => {
            const permission = await Notification.requestPermission();
            console.log('Notification permission:', permission);
        };

        // Call the function to request permission
        requestNotificationPermission();

        connectWS();

        return () => {
            disconnectWS();
        };
    }, []);

    useEffect(() => {
        fetch(`http://${host}/chats`)
            .then(response => response.json())
            .then(data => {
                const result = data.reduce((acc: any, chat: any) => {
                    acc[chat.ID] = chat.Alias;
                    return acc;
                }, {});
                setChats(result);
            });
    }, []);

    const filteredMessages = useMemo(() => {
        if (selectedChat) {
            return messages.filter((message) => message.chat === selectedChat);
        }
        return messages;
    }, [selectedChat, messages]);

    const handleSubmit = async () => {
        await handleUserInteraction();
        setLoading(true);
        setJustOpened(false);
        setLastContent(content);
        fetch(`http://${host}/messages?from=${moment(dateFrom).format('DD.MM.YYYY')}&to=${moment(dateTo).format('DD.MM.YYYY')}&content=${content}`)
            .then(response => response.json())
            .then(data => {
                if (data === null) {
                    setMessages([])
                } else {
                    setMessages(data);
                }
                setLastMessageTs(new Date().toISOString());
            }).finally(() => setLoading(false));
    }

    const handleExportToClipboard = async () => {
        // iterate over messages and get the content if the id is in selectedMessageIds to keep correct order
        const selectedMessages = filteredMessages.filter((message) => selectedMessageIds.includes(message.id));

        const content = selectedMessages.map((message) => message.content).join('\n');

        await navigator.clipboard.writeText(content);
    }

    const handleSelectParsableMessages = () => {
        const selectedIds = filteredMessages.filter((message) => {
            return parseCoordinatesFromContent(message.content) && parseDateTime(message);
        }).map((message) => message.id);
        setSelectedMessageIds(selectedIds);
    }

    const handleExportToJson = async () => {
        const selectedMessages = filteredMessages.filter((message) => selectedMessageIds.includes(message.id));

        const existingContents = new Set<string>();

        const content = selectedMessages.reduce((acc: {
            uuid: string,
            coordinates: [number, number],
            event_at: string,
            note: string
        }[], message) => {
            const parsedDate = parseDateTime(message);
            const parsedCoordinates = parseCoordinatesFromContent(message.content);
            // remove Narrow No-Break Space from the content
            const preparedContent = message.content.replace(/\u202F/g, ' ');
            if (existingContents.has(preparedContent) || !parsedCoordinates || !parsedDate) {
                return acc;
            }
            existingContents.add(preparedContent);
            acc.push({
                uuid: message.id,
                coordinates: parsedCoordinates,
                event_at: parsedDate?.toISOString() || '',
                note: preparedContent,
            });
            return acc;
        }, [])

        const json = JSON.stringify(content, null, 2);
        // await navigator.clipboard.writeText(json);

        // download the file
        const blob = new Blob([json], {type: 'application/json'});
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `messages-${(new Date()).toISOString()}.json`;
        a.click();
        URL.revokeObjectURL(url);
    }

    const handleClose = () => {
        setOpen(false);
    }


    const modalStyle = {
        position: 'absolute',
        top: '50%',
        left: '50%',
        transform: 'translate(-50%, -50%)',
        width: 500,
        bgcolor: 'background.paper',
        border: '2px solid #000',
        boxShadow: 24,
        height: '50vh',
        overflow: 'scroll',
        p: 4,
    };

    const isSelected = (content: string) => {
        if (!selectedContentGroups.length || !content) {
            return false;
        }
        const firstLine = (content.split('\n')[0]).trim();
        return selectedContentGroups.includes(firstLine);
    }

    return (
        <div key={lastMessageTs}>
            <p>{lastMessageTs}</p>
            <Modal
                open={open}
                onClose={handleClose}
            >
                <Box sx={modalStyle}>
                    <div>
                        {contentGroups.map((group) => (
                            <div key={group}>
                                <input type="checkbox" id={group} name={group} value={group}
                                       checked={selectedContentGroups.includes(group)}
                                       onChange={e => {
                                           if (e.target.checked) {
                                               setSelectedContentGroups([...selectedContentGroups, group]);
                                           } else {
                                               setSelectedContentGroups(selectedContentGroups.filter((g) => g !== group));
                                           }
                                       }}/>
                                <label htmlFor={group}>{group}</label>
                            </div>
                        ))}
                    </div>
                </Box>
            </Modal>

            <div className={classes.wrap}>
                <div className={classes.form}>
                    <div className={classes.inputGroupRow}>
                        <div className={classes.inputGroup}>
                            <label>Start date</label>
                            <input type="date" value={moment(dateFrom).format('YYYY-MM-DD')}
                                   onChange={e => setDateFrom(new Date(e.target.value))}/>
                        </div>
                        <div className={classes.inputGroup}>
                            <label>End date</label>
                            <input type="date" value={moment(dateTo).format('YYYY-MM-DD')}
                                   onChange={e => setDateTo(new Date(e.target.value))}/>
                        </div>
                    </div>

                    <div className={classes.inputGroupRow}>

                        <div className={classes.inputGroup}>
                            <label>Content</label>
                            <input type="text" value={content} onChange={e => setContent(e.target.value)}/>
                        </div>
                        <button onClick={handleSubmit}>Search</button>
                        {filteredMessages.length > 0 &&
                            <button onClick={handleSelectParsableMessages}>Select parsable messages</button>}
                        {selectedMessageIds.length > 0 &&
                            <>
                                <button onClick={handleExportToClipboard}>Export to
                                    Clipboard {selectedMessageIds.length} messages
                                </button>
                                <button onClick={handleExportToJson}>Export to JSON</button>
                            </>}
                    </div>

                    {!justOpened && <div className={classes.inputGroupRow}>
                        <label>Pull new messages</label>
                        <input type="checkbox" checked={pullNewMessages}
                               onChange={e => setPullNewMessages(e.target.checked)}/>
                    </div>}

                    {loading ?
                        <div className={classes.spinnerWrap}>
                            <div className={classes.spinner}></div>
                        </div> : <div className={classes.dataTable}>
                            {messages.length ? <table className={classes.table}>
                                <thead>
                                <tr>
                                    <th>
                                        <input type="checkbox" style={{
                                            width: '1.5rem',
                                            height: '1.5rem',
                                            cursor: 'pointer',
                                        }} checked={selectedMessageIds.length === filteredMessages.length}
                                               onChange={(e) => {
                                                   if (e.target.checked) {
                                                       setSelectedMessageIds(filteredMessages.map((message) => message.id));
                                                   } else {
                                                       setSelectedMessageIds([]);
                                                   }
                                               }}/>
                                    </th>
                                    <th>Timestamp</th>
                                    <th>
                                        Chat
                                        <div>
                                            <select name="chat-filter" id="chatFilter"
                                                    onChange={e => setSelectedChat(e.target.value)}>
                                                <option value="">All</option>
                                                {Object.entries(chats).map(([id, alias]) => (
                                                    <option key={id} value={id}>{alias}</option>
                                                ))}
                                            </select>
                                        </div>
                                    </th>
                                    <th>
                                        Content
                                        <div>
                                            <Button onClick={() => {
                                                setOpen(true);
                                            }}>Open filters</Button>
                                        </div>
                                    </th>
                                    <th>Parsed data</th>
                                </tr>
                                </thead>
                                <tbody>
                                {filteredMessages.map((message) => {
                                    //parse string like `2024-08-10 15:06:22 +0300 EEST`
                                    const ts = moment(message.timestamp, 'YYYY-MM-DD HH:mm:ss Z').format('HH:mm:ss DD.MM.YYYY');
                                    let chatName;
                                    if (message.chat in chats) {
                                        chatName = chats[message.chat];
                                    } else {
                                        chatName = message.chat;
                                    }
                                    return (
                                        <tr key={message.id} className={classes.dataRow}>
                                            <td>
                                                <input type="checkbox" style={{
                                                    width: '1.5rem',
                                                    height: '1.5rem',
                                                    cursor: 'pointer',
                                                }} checked={
                                                    selectedMessageIds.includes(message.id)
                                                } onChange={(e) => {
                                                    if (e.target.checked) {
                                                        setSelectedMessageIds([...selectedMessageIds, message.id]);
                                                    } else {
                                                        setSelectedMessageIds(selectedMessageIds.filter((id) => id !== message.id));
                                                    }
                                                }}/>
                                            </td>
                                            <td className={classes.td}>{ts}</td>
                                            <td className={classes.td}>{chatName}</td>
                                            <td className={classes.td}>
                                                <MessageContent message={message} lastContent={lastContent}
                                                                className={isSelected(message.content) ? classes.selected : ''}/>
                                            </td>
                                            <td className={classes.td}>
                                                <ParsedContent message={message}/>
                                            </td>
                                        </tr>
                                    );
                                })}
                                </tbody>
                            </table> : (justOpened ? <p>
                                Press "Search" to get messages
                            </p> : <p>
                                No messages found
                            </p>)}
                        </div>
                    }
                </div>
            </div>
        </div>
    )
}

export default DataViewer;
