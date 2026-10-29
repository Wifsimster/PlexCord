export function stripInlineMarkdown(text) {
    let s = String(text ?? '');
    s = s.replace(/\s*\(\[[^)]*$/, '');
    s = s.replace(/\s*\(\[[^\]]+\]\([^)]*\)\)/g, '');
    s = s.replace(/\[([^\]]+)\]\([^)]*\)/g, '$1');
    s = s.replace(/\*\*([^*]+)\*\*/g, '$1');
    s = s.replace(/\*([^*]+)\*/g, '$1');
    s = s.replace(/`([^`]+)`/g, '$1');
    return s.replace(/\s+/g, ' ').trim();
}

export function parseReleaseNotes(markdown) {
    const sections = [];
    let current = null;

    const ensureSection = () => {
        if (!current) {
            current = { title: '', items: [] };
            sections.push(current);
        }
        return current;
    };

    for (const raw of String(markdown ?? '').split(/\r?\n/)) {
        const line = raw.trim();
        if (!line) continue;

        const heading = line.match(/^(#{1,6})\s+(.*)$/);
        if (heading) {
            if (heading[1].length >= 3) {
                current = { title: stripInlineMarkdown(heading[2]), items: [] };
                sections.push(current);
            }
            continue;
        }

        const bullet = line.match(/^[*+-]\s+(.*)$/);
        const text = stripInlineMarkdown(bullet ? bullet[1] : line);
        if (text) {
            ensureSection().items.push(text);
        }
    }

    return sections.filter((section) => section.items.length > 0);
}

export function summarizeReleaseNotes(markdown, maxLen = 120) {
    const items = parseReleaseNotes(markdown).flatMap((section) => section.items);
    const summary = items.join(' • ');
    if (summary.length <= maxLen) {
        return summary;
    }
    const clipped = summary.slice(0, maxLen);
    const lastSpace = clipped.lastIndexOf(' ');
    return `${(lastSpace > 0 ? clipped.slice(0, lastSpace) : clipped).trimEnd()}…`;
}
