import React, { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import rehypeRaw from 'rehype-raw';
import remarkGfm from 'remark-gfm';
// RemoteMarkdown allows embedding remote markdown documents into the docs.
const RemoteMarkdown = ({ src }) => {
    const [content, setContent] = useState('');

    useEffect(() => {
        fetch(src)
            .then((response) => response.text())
            .then(setContent)
            .catch((error) => console.error('Error fetching markdown:', error));
    }, [src]);

    return <ReactMarkdown
        rehypePlugins={[rehypeRaw]}
        remarkPlugins={[remarkGfm]}
    >{content}</ReactMarkdown>;
};

export default RemoteMarkdown;