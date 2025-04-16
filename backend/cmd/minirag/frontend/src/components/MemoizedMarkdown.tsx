import { marked } from 'marked';
import { memo, useMemo } from 'react';
import ReactMarkdown from 'react-markdown';

function parseMarkdownIntoBlocks(markdown: string): string[] {
  const tokens = marked.lexer(markdown);
  return tokens.map(token => token.raw);
}

const MemoizedMarkdownBlock = memo(
  ({ content }: { content: string }) => <ReactMarkdown>{content}</ReactMarkdown>,
  (prev, next) => prev.content === next.content
);

MemoizedMarkdownBlock.displayName = 'MemoizedMarkdownBlock';

export const MemoizedMarkdown = memo(
  ({ content, id }: { content: string; id: string }) => {
    const blocks = useMemo(() => parseMarkdownIntoBlocks(content), [content]);
    return (
      <div className="markdown-content">
        {blocks.map((block, i) => (
          <MemoizedMarkdownBlock content={block} key={`${id}-block_${i}`} />
        ))}
      </div>
    );
  }
);

MemoizedMarkdown.displayName = 'MemoizedMarkdown'; 