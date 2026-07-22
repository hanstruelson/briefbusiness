export interface Draft {
  id: number;
  title: string;
  markdown: string;
  updated_at: string;
  user_id: number;
}

export interface SearchResult {
  draft: Draft;
  score: number;
  matches: {
    title: boolean;
    content: boolean;
  };
}

/**
 * Searches and ranks drafts based on query terms.
 * Performs a client-side full text match. Matches all terms (AND logic)
 * and ranks title matches higher than body content matches.
 */
export function searchDrafts(drafts: Draft[], query: string): SearchResult[] {
  if (!query || !query.trim()) {
    return drafts.map(d => ({
      draft: d,
      score: 1,
      matches: { title: false, content: false }
    }));
  }

  const queryTerms = query
    .toLowerCase()
    .replace(/[^\w\s]/g, ' ')
    .split(/\s+/)
    .filter(Boolean);

  if (queryTerms.length === 0) {
    return drafts.map(d => ({
      draft: d,
      score: 1,
      matches: { title: false, content: false }
    }));
  }

  const results: SearchResult[] = [];

  for (const draft of drafts) {
    const titleLower = draft.title.toLowerCase();
    const markdownLower = (draft.markdown || '').toLowerCase();

    let score = 0;
    let matchesTitle = false;
    let matchesContent = false;
    let termMatchCount = 0;

    for (const term of queryTerms) {
      let termMatched = false;

      // Match title (weighted high)
      let titleIdx = titleLower.indexOf(term);
      let titleHits = 0;
      while (titleIdx !== -1) {
        titleHits++;
        titleIdx = titleLower.indexOf(term, titleIdx + term.length);
      }
      if (titleHits > 0) {
        score += titleHits * 10;
        matchesTitle = true;
        termMatched = true;
      }

      // Match content body
      let contentIdx = markdownLower.indexOf(term);
      let contentHits = 0;
      while (contentIdx !== -1) {
        contentHits++;
        contentIdx = markdownLower.indexOf(term, contentIdx + term.length);
      }
      if (contentHits > 0) {
        score += contentHits * 1;
        matchesContent = true;
        termMatched = true;
      }

      if (termMatched) {
        termMatchCount++;
      }
    }

    // Require all search terms to match (AND)
    if (termMatchCount === queryTerms.length && score > 0) {
      results.push({
        draft,
        score,
        matches: {
          title: matchesTitle,
          content: matchesContent,
        },
      });
    }
  }

  return results.sort((a, b) => b.score - a.score);
}
