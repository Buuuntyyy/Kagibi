// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

/**
 * Optimises the raster assets served from `public/` to speed up page loads.
 *
 * Strategy (visually lossless):
 *   - Down-scale images that are far larger than their on-screen display size
 *     (e.g. the 2048x2048 logo shown at 36-40px and used as a favicon). Downscaling
 *     to a display-appropriate resolution removes weight with no perceptible loss.
 *   - Re-encode PNGs with maximum lossless compression (no colour quantisation).
 *   - Only overwrite a file when the result is actually smaller.
 *
 * The script reads each file fully into memory before writing, so overwriting the
 * source in place is safe, and it is idempotent (re-running never degrades quality,
 * since it never enlarges and PNG re-encoding is lossless).
 *
 * Run with: npm run optimize:images
 */
import sharp from 'sharp';
import { readFile, writeFile, readdir } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const publicDir = path.join(root, 'public');

// Max on-disk dimension per asset. The logo displays at <=40px (favicon 16-32px);
// avatars display small in the selector/nav. 256px keeps full crispness on 3x
// high-DPI screens while shedding the bulk of the weight.
const targets = [
  { file: path.join(publicDir, 'Logo.png'), maxSize: 256 },
];

const KB = (n) => (n / 1024).toFixed(1) + ' KB';

async function optimizePng(file, maxSize) {
  const input = await readFile(file);
  const meta = await sharp(input).metadata();

  let pipeline = sharp(input, { limitInputPixels: false });
  const needsResize =
    (meta.width && meta.width > maxSize) || (meta.height && meta.height > maxSize);
  if (needsResize) {
    pipeline = pipeline.resize({
      width: maxSize,
      height: maxSize,
      fit: 'inside',
      withoutEnlargement: true,
    });
  }

  // Lossless PNG: no palette quantisation, maximum deflate effort.
  const output = await pipeline
    .png({ compressionLevel: 9, effort: 10, palette: false })
    .toBuffer();

  const rel = path.relative(root, file);
  if (output.length < input.length) {
    await writeFile(file, output);
    const savedPct = (100 * (1 - output.length / input.length)).toFixed(1);
    console.log(
      `✓ ${rel}: ${KB(input.length)} → ${KB(output.length)} (-${savedPct}%)` +
        (needsResize ? ` [${meta.width}x${meta.height} → ${maxSize}px]` : '')
    );
    return { before: input.length, after: output.length };
  }

  console.log(`· ${rel}: already optimal (${KB(input.length)})`);
  return { before: input.length, after: input.length };
}

async function main() {
  const files = [...targets];

  // All avatars share the same treatment.
  const avatarDir = path.join(publicDir, 'avatars');
  try {
    const entries = await readdir(avatarDir);
    for (const name of entries) {
      if (name.toLowerCase().endsWith('.png')) {
        files.push({ file: path.join(avatarDir, name), maxSize: 256 });
      }
    }
  } catch {
    // avatars/ optional
  }

  let totalBefore = 0;
  let totalAfter = 0;
  for (const { file, maxSize } of files) {
    const { before, after } = await optimizePng(file, maxSize);
    totalBefore += before;
    totalAfter += after;
  }

  const savedPct = totalBefore > 0 ? (100 * (1 - totalAfter / totalBefore)).toFixed(1) : '0';
  console.log(
    `\nTotal: ${KB(totalBefore)} → ${KB(totalAfter)} (-${savedPct}%, ` +
      `${KB(totalBefore - totalAfter)} saved)`
  );
}

main().catch((err) => {
  console.error('optimize-images failed:', err);
  process.exit(1);
});
