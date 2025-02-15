import moment from "moment";

import {RawMessage} from "./types.ts";

export const parseDateTime = (message: RawMessage): Date | null => {
    const input = message.content;
    const defaultDate = moment(message.timestamp, 'YYYY-MM-DD HH:mm:ss Z').toDate();

    const patterns = [
        /(?<day>\d{1,2})\.(?<month>\d{1,2})\.(?<year>\d{4}|\d{2})\D*(?<hour>\d{1,2})[:;\.](?<minute>\d{1,2})/, // Full date and time
        /(?<hour>\d{1,2})[:;\.](?<minute>\d{1,2})\D*(?<day>\d{1,2})\.(?<month>\d{1,2})\.(?<year>\d{4}|\d{2})/, // Time before date
        /(?<day>\d{1,2})\.(?<month>\d{1,2})\D*(?<hour>\d{1,2})[:;\.](?<minute>\d{1,2})/, // Date without year
        /(?<hour>\d{1,2})[:;\.](?<minute>\d{1,2})/ // Only time
    ];

    for (const pattern of patterns) {
        const match = input.match(pattern);
        if (match?.groups) {
            const {day, month, year, hour, minute} = match.groups;

            const parsedYear = year ? (year.length === 2 ? 2000 + parseInt(year) : parseInt(year)) : defaultDate.getFullYear();
            const parsedMonth = month ? parseInt(month) - 1 : defaultDate.getMonth();
            const parsedDay = day ? parseInt(day) : defaultDate.getDate();
            const parsedHour = hour ? parseInt(hour) : 0;
            const parsedMinute = minute ? parseInt(minute) : 0;

            return new Date(parsedYear, parsedMonth, parsedDay, parsedHour, parsedMinute);
        }
    }

    return null;
}

export const getYesterdaysDate = () => {
    const date = new Date();
    date.setDate(date.getDate() - 1);
    return date;
}

export const parseCoordinatesFromContent = (content: string): [number, number] | null => {
    const removeNewLines = content.replace(/\n/g, ' ');
    // looking for coordinates in the format like `55.7558,37.6176`
    const coordinatePattern = /(\d+\.\d+),(\d+\.\d+)/g;
    const matches = [...removeNewLines.matchAll(coordinatePattern)];
    if (matches.length) {
        // return [matches[0][1], matches[0][2]];
        // iterate over matches and return the first one that has valid coordinates
        for (const match of matches) {
            const lat = parseFloat(match[1]);
            const lon = parseFloat(match[2]);
            if (lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180) {
                return [lat, lon];
            }
        }
    }
    return null;
}

export const downloadJsonFile = (json: string, filename: string) => {
    const blob = new Blob([json], {type: 'application/json'});
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${filename}-${(new Date()).toISOString()}.json`;
    a.click();
    URL.revokeObjectURL(url);
}

export const uploadJsonFile = (callback: (value: any) => void) => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';
    input.onchange = async (e) => {
        const files = (e.target as HTMLInputElement).files;
        if (files && files.length) {
            const file = files[0];
            const reader = new FileReader();
            reader.onload = async e => {
                if (e.target && e.target.result)
                    try {
                        callback(JSON.parse(e.target.result as string));
                    } catch (error) {
                        if (error instanceof Error) alert('Failed to parse polygons: ' + error.message);
                        else alert('Failed to parse polygons: Unknown error');
                    }
            }
            reader.readAsText(file);
        }
    }
    input.click();
}

export const copyToClipboard = async (content: string) => {
    if (!navigator.clipboard) {
        console.debug('Clipboard API is not available');
        console.debug('Try to use fallback');
        // fallback for browsers that do not support clipboard API
        const textArea = document.createElement('textarea');
        textArea.value = content;
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
        // scroll to the top
        window.scrollTo(0, 0);
    } else {
        await navigator.clipboard.writeText(content);
    }
}

export const isNewNotificationSupported = () => {
    if (!window.Notification || !Notification.requestPermission)
        return false;
    if (Notification.permission == 'granted')
        throw new Error('You must only call this *before* calling Notification.requestPermission(), otherwise this feature detect would bug the user with an actual notification!');
    try {
        new Notification('');
    } catch (e) {
        if ((e as Error).name == 'TypeError'){
            return false;
        }
    }
    return true;
}