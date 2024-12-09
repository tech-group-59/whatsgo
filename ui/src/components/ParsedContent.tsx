import {RawMessage} from "../types";
import moment from "moment/moment";
import {parseCoordinatesFromContent, parseDateTime} from "../helpers.ts";

export const ParsedContent = ({message}: {
    message: RawMessage
}) => {
    const parsedDate = parseDateTime(message);
    const parsedCoordinates = parseCoordinatesFromContent(message.content);
    return (
        <>
            <div>
                {parsedDate ? moment(parsedDate).format('HH:mm DD.MM.YYYY') : ''}
            </div>
            <div>
                {parsedCoordinates ?
                    <a href={`https://www.google.com/maps/search/?api=1&query=${parsedCoordinates[0]},${parsedCoordinates[1]}`}
                       target="_blank"
                       rel="noreferrer">{parsedCoordinates[0].toFixed(2)},{parsedCoordinates[1].toFixed(2)}</a> : ''}

            </div>
        </>
    );
};