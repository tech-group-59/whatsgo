import { LatLngLiteral } from "leaflet";

import { RawMessage } from "./types.ts";
import moment from "moment";

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

export const isPointInPolygon = (point: LatLngLiteral, polygon: LatLngLiteral[]): boolean => {
    let inside = false;
    const { lat, lng } = point;
    const n = polygon.length;

    for (let i = 0, j = n - 1; i < n; j = i++) {
        const xi = polygon[i].lat, yi = polygon[i].lng;
        const xj = polygon[j].lat, yj = polygon[j].lng;

        const intersect = (yi > lng) !== (yj > lng) &&
            (lat < (xj - xi) * (lng - yi) / (yj - yi) + xi);

        if (intersect) inside = !inside;
    }

    return inside;
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
