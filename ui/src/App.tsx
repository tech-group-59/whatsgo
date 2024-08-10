// import { useState } from 'react'
import {createUseStyles} from 'react-jss';
import {useEffect, useState} from "react";

import moment from 'moment';

interface RawMessage {
    id: string;
    sender: string;
    chat: string;
    content: string;
    timestamp: string;
    parsed_content: string;
}

// Define the type for the array of messages
type RawMessages = RawMessage[];


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
    },
    inputGroup: {
        padding: '0 1rem 0 0',
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
        }
    },
    td: {
        padding: '0.5rem',
        'border-bottom': '1px solid #f3f3f3',
    }
})

const getYesterdaysDate = () => {
    const date = new Date();
    date.setDate(date.getDate() - 1);
    return date;
}

function App() {
    const classes = useStyles();
    const [dateFrom, setDateFrom] = useState(getYesterdaysDate());
    const [dateTo, setDateTo] = useState(new Date());
    const [content, setContent] = useState('');
    const [chats, setChats] = useState<{ [key: string]: string }>({});
    const [messages, setMessages] = useState<RawMessages>([]);
    const [loading, setLoading] = useState(false);


    useEffect(() => {
        fetch('/chats')
            .then(response => response.json())
            .then(data => {
                const result = data.reduce((acc: any, chat: any) => {
                    acc[chat.ID] = chat.Alias;
                    return acc;
                }, {});
                setChats(result);
            });
    }, []);


    const handleSubmit = () => {
        setLoading(true);
        fetch(`/messages?from=${moment(dateFrom).format('DD.MM.YYYY')}&to=${moment(dateTo).format('DD.MM.YYYY')}&content=${content}`)
            .then(response => response.json())
            .then(data => {
                if (data === null) {
                    setMessages([]);
                } else {
                    setMessages(data);
                }
            }).finally(() => setLoading(false));
    }

    return (
        <>
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
                    </div>

                    {loading ?
                        <div className={classes.spinnerWrap}>
                            <div className={classes.spinner}></div>
                        </div> : <div className={classes.dataTable}>
                            {messages.length ? <table className={classes.table}>
                                <thead>
                                <tr>
                                    <th>Timestamp</th>
                                    <th>Chat</th>
                                    <th>Content</th>
                                </tr>
                                </thead>
                                <tbody>
                                {messages.map((message) => {
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
                                            <td className={classes.td}>{ts}</td>
                                            <td className={classes.td}>{chatName}</td>
                                            <td className={classes.td}>{message.content}</td>
                                        </tr>
                                    );
                                })}
                                </tbody>
                            </table> : <p>
                                No data. Press "Search" to get messages
                            </p>}
                        </div>
                    }
                </div>
            </div>
        </>
    )
}

export default App
