let decoder;

function initDecoder(func) {
  const decorderInit = {
    output: func,
    error: (e) => {
      console.log(e.message);
    }
  };

  decoder = new VideoDecoder(decorderInit);

  const config = {
    codec: 'avc1.42002a',
    codedWidth: 1920,
    codedHeight: 1080,
    hardwareAcceleration: 'no-preference',
  };
  decoder.configure(config);
}

function inputChunk(data, pts, iskey) {
  const chunk = new EncodedVideoChunk({
    timestamp: pts,
    type: iskey ? 'key' : 'delta',
    data: data
  });
  decoder.decode(chunk);
}