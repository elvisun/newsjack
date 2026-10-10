// Minimal MP4 (ISO BMFF) writer for one H.264 video track, plus a reader for the
// renderer's self-check. Node standard library only.
//
// Input comes straight from Chrome's WebCodecs VideoEncoder:
//   avcC     decoderConfig.description (the encoder was configured with avc: { format: "avc" })
//   samples  encoded chunks in decode order: [{ data: Uint8Array, key: boolean }]
// The stream must be constant frame rate with no B-frames (WebCodecs output timestamps
// are strictly increasing), so no composition offsets or edit lists are needed.
//
// Layout: ftyp, then moov, then mdat ("fast start": players can begin before the whole
// file arrives, as YouTube and LinkedIn recommend).

const u8 = (n) => Buffer.from([n & 255]);
const u16 = (n) => { const b = Buffer.alloc(2); b.writeUInt16BE(n); return b; };
const u32 = (n) => { const b = Buffer.alloc(4); b.writeUInt32BE(n >>> 0); return b; };
const str = (s) => Buffer.from(s, "latin1");
const box = (type, ...parts) => {
  const body = Buffer.concat(parts.flat());
  return Buffer.concat([u32(body.length + 8), str(type), body]);
};
// a "full box" carries a version byte and 24 bits of flags before its body
const fullBox = (type, version, flags, ...parts) =>
  box(type, u8(version), u8(flags >> 16), u16(flags & 0xffff), ...parts);
const IDENTITY_MATRIX = [0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000].map(u32);
const MEDIA_TIMESCALE = 90000;
const MOVIE_TIMESCALE = 1000;

export function muxMp4({ width, height, fps, avcC, samples }) {
  if (!samples.length) throw new Error("muxMp4 needs at least one encoded sample");
  if (width % 2 || height % 2) throw new Error("H.264 needs even frame dimensions");
  if (!samples[0].key) throw new Error("the first sample must be a keyframe");
  const sampleDelta = Math.round(MEDIA_TIMESCALE / fps);
  const count = samples.length;
  const mediaDuration = count * sampleDelta;
  const movieDuration = Math.round((count / fps) * MOVIE_TIMESCALE);

  const build = (mdatPayloadOffset) => {
    const ftyp = box("ftyp", str("isom"), u32(0x200), str("isom"), str("iso2"), str("avc1"), str("mp41"));
    const mvhd = fullBox("mvhd", 0, 0, u32(0), u32(0), u32(MOVIE_TIMESCALE), u32(movieDuration),
      u32(0x10000), u16(0x100), Buffer.alloc(10), IDENTITY_MATRIX, Buffer.alloc(24), u32(2));
    const tkhd = fullBox("tkhd", 0, 3, u32(0), u32(0), u32(1), u32(0), u32(movieDuration), Buffer.alloc(8),
      u16(0), u16(0), u16(0), u16(0), IDENTITY_MATRIX, u32(width << 16), u32(height << 16));
    const mdhd = fullBox("mdhd", 0, 0, u32(0), u32(0), u32(MEDIA_TIMESCALE), u32(mediaDuration), u16(0x55c4), u16(0));
    const hdlr = fullBox("hdlr", 0, 0, u32(0), str("vide"), Buffer.alloc(12), str("VideoHandler\0"));
    const vmhd = fullBox("vmhd", 0, 1, u16(0), u16(0), u16(0), u16(0));
    const dinf = box("dinf", fullBox("dref", 0, 0, u32(1), fullBox("url ", 0, 1)));
    // colr nclx: BT.709 primaries, transfer and matrix, limited range (what canvas frames encode as)
    const colr = box("colr", str("nclx"), u16(1), u16(1), u16(1), u8(0));
    const avc1 = box("avc1", Buffer.alloc(6), u16(1), u16(0), u16(0), Buffer.alloc(12),
      u16(width), u16(height), u32(0x480000), u32(0x480000), u32(0), u16(1),
      Buffer.alloc(32), u16(0x18), u16(0xffff), box("avcC", Buffer.from(avcC)), colr);
    const stsd = fullBox("stsd", 0, 0, u32(1), avc1);
    const stts = fullBox("stts", 0, 0, u32(1), u32(count), u32(sampleDelta));
    const keyIndexes = samples.map((s, i) => (s.key ? i + 1 : 0)).filter(Boolean);
    const stss = fullBox("stss", 0, 0, u32(keyIndexes.length), keyIndexes.map(u32));
    const stsc = fullBox("stsc", 0, 0, u32(1), u32(1), u32(count), u32(1)); // one chunk holds every sample
    const stsz = fullBox("stsz", 0, 0, u32(0), u32(count), samples.map((s) => u32(s.data.length)));
    const stco = fullBox("stco", 0, 0, u32(1), u32(mdatPayloadOffset));
    const stbl = box("stbl", stsd, stts, stss, stsc, stsz, stco);
    const moov = box("moov", mvhd, box("trak", tkhd, box("mdia", mdhd, hdlr, box("minf", vmhd, dinf, stbl))));
    return { ftyp, moov };
  };

  // box sizes do not depend on the offset value, so build once to measure, then for real
  const sized = build(0);
  const payloadOffset = sized.ftyp.length + sized.moov.length + 8;
  const { ftyp, moov } = build(payloadOffset);
  const payload = Buffer.concat(samples.map((s) => Buffer.from(s.data.buffer, s.data.byteOffset, s.data.byteLength)));
  if (payload.length + 8 > 0xffffffff) throw new Error("video is too large for a 32-bit mdat box");
  return Buffer.concat([ftyp, moov, u32(payload.length + 8), str("mdat"), payload]);
}

// Read back the facts a viewer cares about from the boxes we wrote.
export function readMp4Summary(buffer) {
  const children = (start, end) => {
    const out = [];
    for (let at = start; at + 8 <= end;) {
      const size = buffer.readUInt32BE(at);
      if (size < 8 || at + size > end) throw new Error(`malformed box at byte ${at}`);
      out.push({ type: buffer.toString("latin1", at + 4, at + 8), start: at, body: at + 8, end: at + size });
      at += size;
    }
    return out;
  };
  const find = (list, type) => list.find((b) => b.type === type);
  const top = children(0, buffer.length);
  const order = top.map((b) => b.type);
  const moov = find(top, "moov");
  if (!moov) throw new Error("no moov box");
  const trak = find(children(moov.body, moov.end), "trak");
  const mdia = find(children(trak.body, trak.end), "mdia");
  const mdiaKids = children(mdia.body, mdia.end);
  const mdhd = find(mdiaKids, "mdhd");
  const timescale = buffer.readUInt32BE(mdhd.body + 12);
  const duration = buffer.readUInt32BE(mdhd.body + 16);
  const minf = find(mdiaKids, "minf");
  const stbl = find(children(minf.body, minf.end), "stbl");
  const stblKids = children(stbl.body, stbl.end);
  const stsd = find(stblKids, "stsd");
  const entry = children(stsd.body + 8, stsd.end)[0];
  const width = buffer.readUInt16BE(entry.body + 24);
  const height = buffer.readUInt16BE(entry.body + 26);
  const stts = find(stblKids, "stts");
  const sampleDelta = buffer.readUInt32BE(stts.body + 12);
  const stsz = find(stblKids, "stsz");
  const frames = buffer.readUInt32BE(stsz.body + 8);
  const stss = find(stblKids, "stss");
  const keyframes = buffer.readUInt32BE(stss.body + 4);
  return {
    codec: entry.type,
    width,
    height,
    fps: Math.round((timescale / sampleDelta) * 1000) / 1000,
    frames,
    keyframes,
    duration_seconds: Math.round((duration / timescale) * 1000) / 1000,
    faststart: order.indexOf("moov") < order.indexOf("mdat"),
    box_order: order,
  };
}
