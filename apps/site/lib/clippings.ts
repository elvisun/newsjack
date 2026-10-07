import { getTweet, type Tweet } from "react-tweet/api";

// Posts and coverage shown in "the clippings". X posts are fetched at render
// time (cached for an hour) and drawn with X's own embed layout.

// LinkedIn has no stable public API, so these posts are verbatim copies kept
// here and linked to the originals. A leading "…" marks a post trimmed from
// the front.
export type LinkedInPost = {
  url: string;
  author: string;
  authorUrl: string;
  kind: "person" | "company";
  avatar: string | null;
  subtitle?: string;
  publishedAt: string;
  text: string;
  image?: { src: string; width: number; height: number; video?: boolean };
};

const NEWSJACK_PROMO = {
  src: "https://media.licdn.com/dms/image/v2/D5610AQHp0teTanJz5A/image-shrink_1280/B56Z6srStRIoAc-/0/1781013491106?e=2147483647&v=beta&t=fCn5NWexy9kvkZGLjvGznzhzS4n44bNN3TQF8fnDVDo",
  width: 825,
  height: 1036,
};

export const LINKEDIN_POSTS: Record<string, LinkedInPost> = {
  ryan: {
    url: "https://www.linkedin.com/posts/ryanmccormick131_pr-publicrelations-activity-7513258675474579456-6t3S",
    author: "Ryan McCormick",
    authorUrl: "https://www.linkedin.com/in/ryanmccormick131",
    kind: "person",
    avatar: "/clippings/ryan.jpg",
    publishedAt: "2026-10-06",
    text: "I think Elvis Sun is a genius. Highly recommend following him and checking out his new site. #PR #PublicRelations",
  },
  julie: {
    url: "https://www.linkedin.com/posts/juliekonners_gamechanger-one-thing-ive-said-for-a-while-activity-7475585174450352128-_jfV",
    author: "Julie Konners Handler",
    authorUrl: "https://www.linkedin.com/in/juliekonners",
    kind: "person",
    avatar:
      "https://media.licdn.com/dms/image/v2/C5603AQH1zFL1uuEuag/profile-displayphoto-shrink_200_200/profile-displayphoto-shrink_200_200/0/1626449123571?e=2147483647&v=beta&t=eoGlj_uzxcXt2GyhqvIbvOibvh1A6tzy3vLj7Po5cGw",
    publishedAt: "2026-06-24",
    text: "Gamechanger.\n\nOne thing I've said for a while is that AI still hasn't really cracked press clips.\n\nNot because generating a client-ready PDF is hard, but because building a good clip actually takes time. We do this for every placement, across every client, every day.\n\nA few weeks ago, I mentioned that to Elvis Sun as he walked me through Newsjack, an open-source library of AI skills built specifically for PR. Pretty cool to see it turn into an actual tool.\n\nIf this works the way it looks, it could save PR teams a ridiculous amount of time on one of the most manual parts of the job.\n\nMore of this, please. Less replacing the work. More making the work easier.",
  },
  digitalpr: {
    url: "https://www.linkedin.com/posts/wearedigitalpr_digitalpr-aiforpr-prskills-activity-7486428577274499072-iE3F",
    author: "We Are DigitalPR",
    authorUrl: "https://www.linkedin.com/company/wearedigitalpr",
    kind: "company",
    avatar:
      "https://media.licdn.com/dms/image/v2/D4E0BAQFmSy2b8JI5Ew/company-logo_100_100/B4EZ.e4LefJUAI-/0/1785076955302/wearedigitalpr_logo?e=2147483647&v=beta&t=-OG8DxyYvxy24lVK2evimIk4V1nPrxw5hCsYyk3kSeM",
    publishedAt: "2026-07-24",
    text: "Newsjack is one of the most exciting open-source PR resources we've seen in a long time.\n\nOne of our community members, Elvis Sun has made 18 AI-powered PR skills completely free and open source, helping founders, in-house teams, freelancers, and agencies access capabilities that were traditionally locked behind expensive PR software.\n\nFrom identifying newsjacking opportunities and tracking media coverage to generating pitch angles, finding the right journalists, fact-checking claims, and even stress-testing your pitches with a \"meanest editor,\" it's essentially an AI-powered PR team built for modern communicators.\n\nWhat we particularly like is the strong ethical framework behind it. It actively discourages spammy outreach, fabricated quotes, and poor PR practices.\n\nIf you're interested in reactive PR, AI workflows, or simply want to experiment with how AI can augment PR skills, this is definitely worth exploring.\n\nA fantastic example of how AI is democratizing access to professional PR capabilities.\n\n#digitalpr #aiforpr #prskills",
    image: {
      src: "https://media.licdn.com/dms/image/v2/D4E22AQE7d8v3n6RzYA/feedshare-shrink_1280/B4EZ.UjHlqH4AQ-/0/1784903662274?e=2147483647&v=beta&t=n4ZfLOLWR_xcIjkoYsuLdZSULVvUy459HvaWS01o00M",
      width: 825,
      height: 1036,
    },
  },
  victoria: {
    url: "https://www.linkedin.com/posts/andriiashkina_i-came-across-a-great-piece-by-lindsay-bennett-activity-7510435724991827968-xDH0",
    author: "Victoria Andriiashkina",
    authorUrl: "https://www.linkedin.com/in/andriiashkina",
    kind: "person",
    avatar: "/clippings/victoria.jpg",
    subtitle:
      "Global Public Relations Manager | SaaS, B2B/B2C tech, climate & sustainability | Earned media in Forbes, WIRED, TechCrunch, TechRadar",
    publishedAt: "2026-09-28",
    text: "\u2026 And one more thing I\u2019d add from my own experience: I also love the open source PR skill built by Elvis Sun. You can install it in Claude and use it as an AI powered PR assistant for a range of communications tasks. I use it regularly and it has become one of those things that genuinely makes my workflow easier. Highly recommend checking it out.",
  },
  michelle: {
    url: "https://www.linkedin.com/posts/michellekirkandrade_setting-this-up-now-and-cant-wait-to-try-activity-7472336377695711232-ri70",
    author: "Michelle Andrade",
    authorUrl: "https://www.linkedin.com/in/michellekirkandrade",
    kind: "person",
    avatar:
      "https://media.licdn.com/dms/image/v2/D5603AQFxACJXVPoU6w/profile-displayphoto-shrink_200_200/profile-displayphoto-shrink_200_200/0/1729634616270?e=2147483647&v=beta&t=J9SxOvde9z4meHmGgcQGM4b_X8kmvgCA9jgvUceJes8",
    publishedAt: "2026-06-15",
    text: "Setting this up now and can't wait to try it; thanks, Elvis Sun! \u26a1\ufe0f",
    image: NEWSJACK_PROMO,
  },
  ary: {
    url: "https://www.linkedin.com/posts/aryaranguiz_smallbusiness-activity-7472338939509571585-qwmu",
    author: "Ary Aranguiz, CPTD, AIPM",
    authorUrl: "https://www.linkedin.com/in/aryaranguiz",
    kind: "person",
    avatar: null,
    publishedAt: "2026-06-15",
    text: "#Smallbusiness owners, if you didn't have the budget to hire a Public Relations firm, now you can build your own PR team with Claude Cowork. Check out Elvis Sun's 18 PR Skills for building and amplyifying your company brand!",
  },
  futureReady: {
    url: "https://www.linkedin.com/posts/future-ready-ai-group_smallbusiness-activity-7472339175430856704-Zq1T",
    author: "Future-Ready AI Group, LLC",
    authorUrl: "https://www.linkedin.com/company/future-ready-ai-group",
    kind: "company",
    avatar:
      "https://media.licdn.com/dms/image/v2/D4E0BAQHNnta9x0V8dA/company-logo_100_100/B4EZWcmSlCHgAY-/0/1742089038535/future_ready_ai_group_logo?e=2147483647&v=beta&t=g2ESt7PV4l9JK8OGydPE66My9ezmxfXkSKk_R3futks",
    subtitle: "681 followers",
    publishedAt: "2026-06-15",
    text: "#Smallbusiness owners, if you didn't have the budget to hire a Public Relations firm, now you can build your own PR team with Claude Cowork. Check out Elvis Sun's 18 PR Skills for building and amplyifying your company brand!",
    image: NEWSJACK_PROMO,
  },
};

// The wall, in display order. X and LinkedIn alternate so neither platform
// clumps. Add a post by adding one line.
export type Clipping =
  { x: string } | { linkedin: string } | { youtube: string };

// YouTube videos that cover newsjack, starting at the segment about it.
export const YOUTUBE_VIDEOS: Record<
  string,
  {
    id: string;
    start: number;
    title: string;
    channel: string;
    channelUrl: string;
    channelAvatar: string;
    publishedAt: string;
    chapter: string;
  }
> = {
  nextNewThing: {
    id: "ll16Bvyw54A",
    start: 180,
    title: "Get paid to use Claude, get publicity, and more free ai tools",
    channel: "The Next New Thing",
    channelUrl: "https://www.youtube.com/@TheNextNewThingAI",
    channelAvatar:
      "https://yt3.ggpht.com/ubw-qu3I8u2GG27Cw8TRwCshGU6YdB2bL2oBKMvXEiWs58-ziD8SYXaw180RTsV7vGQ7lF4o8A=s88-c-k-c0x00ffffff-no-rj",
    publishedAt: "2026-06-18",
    chapter: "NewsJack.sh (AI-powered PR outreach agent)",
  },
};

export const CLIPPINGS: Clipping[] = [
  { x: "2064360727184384476" }, // Connor Showler, quoting the launch post
  { linkedin: "victoria" },
  { youtube: "nextNewThing" },
  { x: "2100951347080421409" }, // Elvis: Jev demo
  { linkedin: "julie" },
  { x: "2104315895669789084" }, // Every
  { x: "2064428394390175881" }, // Bei Zhang
  { linkedin: "digitalpr" },
  { x: "2064528420009320683" }, // Xghost (Chinese)
  { linkedin: "ryan" },
  { x: "2065434256869065128" }, // Elvis: headline-generator
  { linkedin: "michelle" },
  { x: "2101117097728471219" }, // Fukku (Japanese)
  { x: "2064344412335837203" }, // Nic Polotnianko
  { linkedin: "ary" },
  { x: "2064704921677730108" }, // Elvis: Fable vs Opus
  { x: "2101157437143408800" }, // Leo Lu
  { linkedin: "futureReady" },
];

// English translations for non-English X posts, shown behind a "Show
// translation" toggle. Static text; no runtime translation service.
export const TRANSLATIONS: Record<string, { language: string; text: string }> =
  {
    "2101117097728471219": {
      language: "Japanese",
      text: "A bigger shock than when Claude Code or Astra first came out.\n\nIt really shows how insane Jev is.",
    },
    "2064528420009320683": {
      language: "Chinese",
      text: "What's really interesting isn't that \"agents can write PR copy.\"\n\nIt's that marketing is turning into a loop you can run: watch the news and market signals, generate angles, match journalists and channels, and record which narratives actually earned exposure.\n\nThat reprices a small team's growth ability, from inspiration and connections, into systematized distribution memory.",
    },
  };

export async function loadTweet(id: string): Promise<Tweet | null> {
  try {
    return (await getTweet(id, { next: { revalidate: 3600 } })) ?? null;
  } catch {
    return null;
  }
}

type CardValue = {
  string_value?: string;
  image_value?: { url: string };
};

export type LinkPreview = {
  url: string;
  domain: string;
  title: string;
  image: string;
};

// react-tweet does not draw link previews; read the summary card ourselves.
export function linkPreview(tweet: Tweet): LinkPreview | null {
  const values = (
    tweet as Tweet & { card?: { binding_values?: Record<string, CardValue> } }
  ).card?.binding_values;
  const url = values?.card_url?.string_value;
  const image =
    values?.summary_photo_image_large?.image_value?.url ??
    values?.thumbnail_image_large?.image_value?.url;
  if (!values || !url || !image) return null;
  return {
    url,
    domain:
      values.vanity_url?.string_value ?? values.domain?.string_value ?? "",
    title: values.title?.string_value ?? "",
    image,
  };
}
