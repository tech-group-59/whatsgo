export const HighlightText = ({text, highlight}: {
    text: string,
    highlight: string
}) => {
    const escapedHighlight = highlight.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

    const parts = text.split(new RegExp(`(${escapedHighlight})`, 'gi'));

    if (!highlight) {
        return <span>{text}</span>;
    }

    return (
        <span>
      {parts.map((part, index) => (
          part.toLowerCase() === highlight.toLowerCase() ?
              <mark key={index}>{part}</mark> :
              part
      ))}
    </span>
    );
};