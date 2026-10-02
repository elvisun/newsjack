import {
  Globe,
  MessageSquare,
  Play,
  Repeat2,
  Send,
  ThumbsUp,
  UserRound,
} from "lucide-react";
import Image from "next/image";
import {
  QuotedTweet,
  QuotedTweetBody,
  QuotedTweetContainer,
  QuotedTweetHeader,
  TweetActions,
  TweetBody,
  TweetContainer,
  TweetHeader,
  TweetInReplyTo,
  TweetInfo,
  TweetMedia,
  TweetReplies,
  enrichTweet,
} from "react-tweet";
import type { Tweet } from "react-tweet/api";

import {
  CLIPPINGS,
  LINKEDIN_POSTS,
  TRANSLATIONS,
  YOUTUBE_VIDEOS,
  linkPreview,
  loadTweet,
  type LinkPreview,
  type LinkedInPost,
} from "../../lib/clippings";
import { SectionHeading } from "./section-heading";
import { Translatable } from "./translatable";
import { YouTubeCard } from "./youtube";

// X's summary card: rounded image with the title overlaid, domain below.
function XLinkCard({ preview }: { preview: LinkPreview }) {
  return (
    <a
      className="mt-3 block"
      href={preview.url}
      rel="noreferrer"
      target="_blank"
    >
      <span className="relative block aspect-[1.91/1] overflow-hidden rounded-2xl border border-[#cfd9de]">
        <Image
          alt=""
          className="object-cover"
          fill
          sizes="(min-width: 1280px) 400px, (min-width: 768px) 50vw, 100vw"
          src={preview.image}
          unoptimized
        />
        {preview.title && (
          <span className="absolute bottom-3 left-3 max-w-[85%] truncate rounded bg-black/75 px-1.5 text-[13px] leading-5 text-white">
            {preview.title}
          </span>
        )}
      </span>
      <span className="mt-1 block text-[13px] text-[#536471]">
        From {preview.domain}
      </span>
    </a>
  );
}

// X's own embed layout (react-tweet), plus the link preview it doesn't draw.
// A compact quote drops the quoted post's media when it already shows
// elsewhere on the wall.
function XCard({
  tweet,
  compactQuote = false,
}: {
  tweet: Tweet;
  compactQuote?: boolean;
}) {
  const full = enrichTweet(tweet);
  const quotedId = tweet.quoted_tweet?.id_str;
  // X hides the link to the quoted post in the text; so do we.
  const enriched = quotedId
    ? {
        ...full,
        entities: full.entities.filter(
          (entity) =>
            !(
              entity.type === "url" &&
              entity.href.includes(`/status/${quotedId}`)
            ),
        ),
      }
    : full;
  const preview = enriched.mediaDetails?.length ? null : linkPreview(tweet);

  return (
    <TweetContainer>
      <TweetHeader tweet={enriched} />
      {enriched.in_reply_to_status_id_str && (
        <TweetInReplyTo tweet={enriched} />
      )}
      {TRANSLATIONS[tweet.id_str] ? (
        <Translatable
          language={TRANSLATIONS[tweet.id_str].language}
          original={<TweetBody tweet={enriched} />}
          translation={TRANSLATIONS[tweet.id_str].text}
        />
      ) : (
        <TweetBody tweet={enriched} />
      )}
      {enriched.mediaDetails?.length ? <TweetMedia tweet={enriched} /> : null}
      {preview && <XLinkCard preview={preview} />}
      {enriched.quoted_tweet &&
        (compactQuote ? (
          <QuotedTweetContainer tweet={enriched.quoted_tweet}>
            <QuotedTweetHeader tweet={enriched.quoted_tweet} />
            <QuotedTweetBody tweet={enriched.quoted_tweet} />
          </QuotedTweetContainer>
        ) : (
          <QuotedTweet tweet={enriched.quoted_tweet} />
        ))}
      <TweetInfo tweet={enriched} />
      <TweetActions tweet={enriched} />
      <TweetReplies tweet={enriched} />
    </TweetContainer>
  );
}

function LinkedInLogo() {
  return (
    <svg
      aria-label="LinkedIn"
      height="21"
      role="img"
      viewBox="0 0 24 24"
      width="21"
    >
      <path
        d="M20.45 20.45h-3.56v-5.57c0-1.33-.02-3.04-1.85-3.04-1.85 0-2.14 1.45-2.14 2.94v5.67H9.35V9h3.41v1.56h.05c.48-.9 1.64-1.85 3.37-1.85 3.6 0 4.27 2.37 4.27 5.46v6.28ZM5.34 7.43a2.06 2.06 0 1 1 0-4.13 2.06 2.06 0 0 1 0 4.13ZM7.12 20.45H3.56V9h3.56v11.45ZM22.22 0H1.77C.79 0 0 .77 0 1.73v20.54C0 23.23.79 24 1.77 24h20.45c.98 0 1.78-.77 1.78-1.73V1.73C24 .77 23.2 0 22.22 0Z"
        fill="#0A66C2"
      />
    </svg>
  );
}

// LinkedIn's short relative age: 3d, 2w, 2mo, 1yr.
function linkedInAge(iso: string, now = new Date()): string {
  const days = Math.max(
    0,
    Math.floor((now.getTime() - new Date(iso).getTime()) / 86_400_000),
  );
  if (days < 7) return `${days}d`;
  if (days < 30) return `${Math.floor(days / 7)}w`;
  if (days < 365) return `${Math.floor(days / 30)}mo`;
  return `${Math.floor(days / 365)}yr`;
}

const LINKEDIN_ACTIONS = [
  { label: "Like", icon: ThumbsUp },
  { label: "Comment", icon: MessageSquare },
  { label: "Repost", icon: Repeat2 },
  { label: "Send", icon: Send },
];

function LinkedInAvatar({ post }: { post: LinkedInPost }) {
  const shape = post.kind === "person" ? "rounded-full" : "rounded";
  if (!post.avatar) {
    return (
      <span
        className={`flex size-12 items-center justify-center bg-[#e9e5df] text-[#a7a29a] ${shape}`}
      >
        <UserRound aria-hidden="true" size={28} />
      </span>
    );
  }
  return (
    <Image
      alt=""
      className={`size-12 object-cover ${shape}`}
      height={48}
      src={post.avatar}
      unoptimized
      width={48}
    />
  );
}

// A LinkedIn post embed, drawn to match LinkedIn's own card.
function LinkedInCard({ post }: { post: LinkedInPost }) {
  return (
    <article
      className="w-full max-w-[550px] overflow-hidden rounded-lg bg-white text-[14px] leading-5 text-black/90 shadow-[0_0_0_1px_rgba(140,140,140,0.2)]"
      style={{
        fontFamily:
          '-apple-system, system-ui, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
      }}
    >
      <header className="flex items-start gap-2 px-4 pt-3">
        <a
          className="shrink-0"
          href={post.authorUrl}
          rel="noreferrer"
          target="_blank"
        >
          <LinkedInAvatar post={post} />
        </a>
        <div className="min-w-0 flex-1">
          <a
            className="block truncate font-semibold hover:text-[#0A66C2] hover:underline"
            href={post.authorUrl}
            rel="noreferrer"
            target="_blank"
          >
            {post.author}
          </a>
          {post.subtitle && (
            <p className="truncate text-xs text-black/60">{post.subtitle}</p>
          )}
          <p className="flex items-center gap-1 text-xs text-black/60">
            {linkedInAge(post.publishedAt)} •
            <Globe aria-label="Visible to anyone" size={12} />
          </p>
        </div>
        <a href={post.url} rel="noreferrer" target="_blank">
          <LinkedInLogo />
        </a>
      </header>

      <div className="px-4 pt-2 pb-2">
        <p
          className={`whitespace-pre-line ${post.image ? "line-clamp-6" : "line-clamp-[10]"}`}
        >
          {post.text.split(/(#\w+)/).map((part, index) =>
            part.startsWith("#") ? (
              <span className="font-semibold text-[#0A66C2]" key={index}>
                {part}
              </span>
            ) : (
              <span key={index}>{part}</span>
            ),
          )}
        </p>
        <a
          className="block text-right text-black/60 hover:text-[#0A66C2] hover:underline"
          href={post.url}
          rel="noreferrer"
          target="_blank"
        >
          …more
        </a>
      </div>

      {post.image && (
        <a
          aria-label={
            post.image.video ? "Watch the video on LinkedIn" : undefined
          }
          className="relative block"
          href={post.url}
          rel="noreferrer"
          target="_blank"
        >
          <Image
            alt=""
            className={`w-full object-cover object-top ${
              post.image.video ? "aspect-[4/5]" : "aspect-[1.91/1]"
            }`}
            height={post.image.height}
            src={post.image.src}
            unoptimized
            width={post.image.width}
          />
          {post.image.video && (
            <span className="absolute inset-0 flex items-center justify-center">
              <span className="flex size-14 items-center justify-center rounded-full bg-black/60 text-white ring-2 ring-white/80">
                <Play
                  aria-hidden="true"
                  className="ml-0.5 fill-white"
                  size={22}
                />
              </span>
            </span>
          )}
        </a>
      )}

      <nav className="mx-4 flex justify-between border-t border-black/10 py-1">
        {LINKEDIN_ACTIONS.map(({ label, icon: Icon }) => (
          <a
            className="flex items-center gap-1.5 rounded px-2 py-3 text-[13px] font-semibold text-black/60 hover:bg-black/5"
            href={post.url}
            key={label}
            rel="noreferrer"
            target="_blank"
          >
            <Icon aria-hidden="true" size={18} />
            <span className="hidden sm:inline">{label}</span>
          </a>
        ))}
      </nav>
    </article>
  );
}

export async function Clippings({ number }: { number: string }) {
  const xIds = CLIPPINGS.flatMap((item) => ("x" in item ? [item.x] : []));
  const tweets = new Map(
    await Promise.all(
      xIds.map(async (id) => [id, await loadTweet(id)] as const),
    ),
  );

  // A quoted post shows its media once: never when it has its own card, and
  // only on the first card that quotes it.
  const shown = new Set(xIds);
  const cards = CLIPPINGS.map((item) => {
    if ("youtube" in item) {
      const video = YOUTUBE_VIDEOS[item.youtube];
      return video ? <YouTubeCard key={item.youtube} {...video} /> : null;
    }
    if ("linkedin" in item) {
      const post = LINKEDIN_POSTS[item.linkedin];
      return post ? <LinkedInCard key={item.linkedin} post={post} /> : null;
    }
    const tweet = tweets.get(item.x);
    if (!tweet) return null;
    const quotedId = tweet.quoted_tweet?.id_str;
    const compactQuote = quotedId ? shown.has(quotedId) : false;
    if (quotedId) shown.add(quotedId);
    return <XCard compactQuote={compactQuote} key={item.x} tweet={tweet} />;
  }).filter(Boolean);

  return (
    <section aria-labelledby="clippings" className="px-6 py-32">
      <div className="mx-auto max-w-7xl">
        <SectionHeading
          id="clippings"
          number={number}
          subtitle="What people are saying about newsjack, and what we've been shipping."
          title="the clippings"
        />
        <div
          className="nj-clippings columns-1 gap-6 md:columns-2 xl:columns-3"
          data-theme="light"
        >
          {cards.map((card, index) => (
            <div
              className="mb-6 flex break-inside-avoid justify-center"
              key={index}
            >
              {card}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
