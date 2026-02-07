import { FFmpeg } from '@ffmpeg/ffmpeg'
import { fetchFile, toBlobURL } from '@ffmpeg/util'

let ffmpegInstance: FFmpeg | null = null
let loadPromise: Promise<FFmpeg> | null = null

export interface VideoTrimOptions {
  startTime: number
  endTime: number
}

export interface VideoMergeOptions {
  clips: Array<{
    url: string
    startTime?: number
    endTime?: number
  }>
}

export interface AudioClip {
  url: string
  startTime: number
  endTime?: number
  duration?: number
  position: number
  volume?: number
}

export interface ProgressCallback {
  (progress: number): void
}

async function getFFmpeg(): Promise<FFmpeg> {
  if (ffmpegInstance) {
    return ffmpegInstance
  }

  if (loadPromise) {
    return loadPromise
  }

  loadPromise = (async () => {
    const ffmpeg = new FFmpeg()

    ffmpeg.on('log', ({ message }) => {
      console.log('[FFmpeg]', message)
    })

    const baseURL = 'https://unpkg.com/@ffmpeg/core@0.12.6/dist/umd'
    await ffmpeg.load({
      coreURL: await toBlobURL(`${baseURL}/ffmpeg-core.js`, 'text/javascript'),
      wasmURL: await toBlobURL(`${baseURL}/ffmpeg-core.wasm`, 'application/wasm')
    })

    ffmpegInstance = ffmpeg
    return ffmpeg
  })()

  return loadPromise
}

export async function trimVideo(
  videoUrl: string,
  options: VideoTrimOptions,
  onProgress?: ProgressCallback
): Promise<Blob> {
  const ffmpeg = await getFFmpeg()

  if (onProgress) onProgress(10)

  const inputFileName = 'input.mp4'
  const outputFileName = 'output.mp4'

  await ffmpeg.writeFile(inputFileName, await fetchFile(videoUrl))

  if (onProgress) onProgress(30)

  const args = [
    '-i', inputFileName,
    '-ss', options.startTime.toString(),
    '-to', options.endTime.toString(),
    '-c', 'copy',
    '-avoid_negative_ts', '1',
    outputFileName
  ]

  await ffmpeg.exec(args)

  if (onProgress) onProgress(80)

  const data = await ffmpeg.readFile(outputFileName) as Uint8Array

  await ffmpeg.deleteFile(inputFileName)
  await ffmpeg.deleteFile(outputFileName)

  if (onProgress) onProgress(100)

  return new Blob([new Uint8Array(data)], { type: 'video/mp4' })
}

export async function mergeVideos(
  options: VideoMergeOptions,
  onProgress?: ProgressCallback
): Promise<Blob> {
  const ffmpeg = await getFFmpeg()

  if (onProgress) onProgress(5)

  const tempFiles: string[] = []
  
  for (let i = 0; i < options.clips.length; i++) {
    const clip = options.clips[i]
    const fileName = `clip_${i}.mp4`
    
    await ffmpeg.writeFile(fileName, await fetchFile(clip.url))
    tempFiles.push(fileName)
    
    if (onProgress) {
      onProgress(5 + (i + 1) / options.clips.length * 40)
    }
  }

  const listContent = tempFiles.map(file => `file '${file}'`).join('\n')
  await ffmpeg.writeFile('filelist.txt', new TextEncoder().encode(listContent))

  if (onProgress) onProgress(50)

  await ffmpeg.exec([
    '-f', 'concat',
    '-safe', '0',
    '-i', 'filelist.txt',
    '-c', 'copy',
    'output.mp4'
  ])

  if (onProgress) onProgress(90)

  const data = await ffmpeg.readFile('output.mp4') as Uint8Array

  for (const file of tempFiles) {
    await ffmpeg.deleteFile(file)
  }
  await ffmpeg.deleteFile('filelist.txt')
  await ffmpeg.deleteFile('output.mp4')

  if (onProgress) onProgress(100)

  return new Blob([new Uint8Array(data)], { type: 'video/mp4' })
}

export async function trimAndMergeVideos(
  clips: Array<{
    url: string
    startTime: number
    endTime: number
  }>,
  onProgress?: ProgressCallback,
  audioClips: AudioClip[] = []
): Promise<Blob> {
  const ffmpeg = await getFFmpeg()

  if (onProgress) onProgress(5)

  const trimmedFiles: string[] = []
  
  for (let i = 0; i < clips.length; i++) {
    const clip = clips[i]
    const inputName = `input_${i}.mp4`
    const outputName = `trimmed_${i}.mp4`
    
    await ffmpeg.writeFile(inputName, await fetchFile(clip.url))
    
    await ffmpeg.exec([
      '-i', inputName,
      '-ss', clip.startTime.toString(),
      '-to', clip.endTime.toString(),
      '-c', 'copy',
      '-avoid_negative_ts', '1',
      outputName
    ])
    
    await ffmpeg.deleteFile(inputName)
    trimmedFiles.push(outputName)
    
    if (onProgress) {
      onProgress(5 + (i + 1) / clips.length * 60)
    }
  }

  const listContent = trimmedFiles.map(file => `file '${file}'`).join('\n')
  await ffmpeg.writeFile('filelist.txt', new TextEncoder().encode(listContent))

  if (onProgress) onProgress(70)

  await ffmpeg.exec([
    '-f', 'concat',
    '-safe', '0',
    '-i', 'filelist.txt',
    '-c', 'copy',
    'final.mp4'
  ])

  if (onProgress) onProgress(95)

  let outputFile = 'final.mp4'
  if (audioClips.length > 0) {
    outputFile = await mixAudioWithVideo(ffmpeg, audioClips, outputFile)
  }

  const data = await ffmpeg.readFile(outputFile) as Uint8Array

  for (const file of trimmedFiles) {
    await ffmpeg.deleteFile(file)
  }
  await ffmpeg.deleteFile('filelist.txt')
  await ffmpeg.deleteFile('final.mp4')
  if (outputFile !== 'final.mp4') {
    await ffmpeg.deleteFile(outputFile)
  }

  if (onProgress) onProgress(100)

  return new Blob([new Uint8Array(data)], { type: 'video/mp4' })
}

async function hasAudioStream(ffmpeg: FFmpeg, fileName: string): Promise<boolean> {
  try {
    await ffmpeg.exec([
      '-i', fileName,
      '-map', '0:a:0',
      '-f', 'null',
      '-'
    ])
    return true
  } catch {
    return false
  }
}

async function mixAudioWithVideo(
  ffmpeg: FFmpeg,
  audioClips: AudioClip[],
  inputFile: string
): Promise<string> {
  const validClips = audioClips.filter(clip => clip.url)
  if (validClips.length === 0) return inputFile

  const audioInputs: Array<{ fileName: string; clip: AudioClip }> = []
  for (let i = 0; i < validClips.length; i++) {
    const clip = validClips[i]
    const extMatch = clip.url.split('?')[0]?.match(/\.(\w{1,5})$/)
    const ext = extMatch ? `.${extMatch[1]}` : '.mp3'
    const audioFileName = `audio_${i}${ext}`
    await ffmpeg.writeFile(audioFileName, await fetchFile(clip.url))
    audioInputs.push({ fileName: audioFileName, clip })
  }

  const inputs: string[] = ['-i', inputFile]
  audioInputs.forEach(input => {
    inputs.push('-i', input.fileName)
  })

  const filterParts: string[] = []
  const mixInputs: string[] = []

  if (await hasAudioStream(ffmpeg, inputFile)) {
    filterParts.push('[0:a]asetpts=PTS-STARTPTS[basea]')
    mixInputs.push('[basea]')
  }

  audioInputs.forEach((input, index) => {
    const clip = input.clip
    const inputIndex = index + 1
    const label = `a${index}`
    const startTime = clip.startTime || 0
    const duration = clip.duration ?? ((clip.endTime ?? 0) - startTime)
    const volume = clip.volume && clip.volume > 0 ? clip.volume : 1
    const delayMs = Math.max(0, Math.round((clip.position || 0) * 1000))

    let filter = `[${inputIndex}:a]`
    if (startTime > 0 || duration > 0) {
      filter += `atrim=start=${startTime}`
      if (duration > 0) {
        filter += `:duration=${duration}`
      }
      filter += ','
    }
    filter += 'asetpts=PTS-STARTPTS'
    if (volume !== 1) {
      filter += `,volume=${volume}`
    }
    if (delayMs > 0) {
      filter += `,adelay=${delayMs}:all=1`
    }
    filter += `[${label}]`

    filterParts.push(filter)
    mixInputs.push(`[${label}]`)
  })

  if (mixInputs.length === 0) {
    for (const input of audioInputs) {
      await ffmpeg.deleteFile(input.fileName)
    }
    return inputFile
  }

  filterParts.push(`${mixInputs.join('')}amix=inputs=${mixInputs.length}:normalize=0,apad[aout]`)
  const filterComplex = filterParts.join(';')

  const outputFile = 'final_audio.mp4'
  await ffmpeg.exec([
    ...inputs,
    '-filter_complex', filterComplex,
    '-map', '0:v',
    '-map', '[aout]',
    '-c:v', 'copy',
    '-c:a', 'aac',
    '-b:a', '128k',
    '-shortest',
    '-y',
    outputFile
  ])

  for (const input of audioInputs) {
    await ffmpeg.deleteFile(input.fileName)
  }

  return outputFile
}

export async function isFFmpegLoaded(): Promise<boolean> {
  return ffmpegInstance !== null
}

export async function unloadFFmpeg(): Promise<void> {
  if (ffmpegInstance) {
    await ffmpegInstance.terminate()
    ffmpegInstance = null
    loadPromise = null
  }
}
