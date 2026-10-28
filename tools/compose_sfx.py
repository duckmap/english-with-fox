import numpy as np,wave,os
from pathlib import Path
root=Path(__file__).resolve().parents[1]
sr=44100
for name,duration,seq,brass in [('win',2.8,[(72,0,.20),(72,.24,.20),(76,.48,.28),(79,.8,.32),(84,1.16,1.0),(72,1.16,1.0),(76,1.16,1.0),(79,1.16,1.0)],True),('try_again',2.1,[(76,0,.35),(74,.34,.35),(72,.68,.40),(67,1.10,.30),(72,1.44,.50)],False)]:
 a=np.zeros(int(sr*duration));
 for midi,start,length in seq:
  t=np.arange(int(length*sr))/sr;f=440*2**((midi-69)/12)
  if brass:tone=sum(np.sin(2*np.pi*f*k*t)/k**1.65 for k in range(1,7));env=np.minimum(t/.025,1)*np.minimum((length-t)/.14,1)*(1-.25*t/length)
  else:tone=np.sin(2*np.pi*f*t)+.12*np.sin(4*np.pi*f*t);env=np.minimum(t/.01,1)*np.exp(-t*4)*np.minimum((length-t)/.06,1)
  v=tone*env;idx=int(start*sr);a[idx:idx+len(v)]+=.20*v
 # short quiet delay gives the fanfare space without harsh peaks
 delay=int(.09*sr);a[delay:]+=.15*a[:-delay];a*=.70/max(abs(a));st=np.stack([a,a*.97],axis=1)
 with wave.open(str(root/'assets/sfx'/(name+'.wav')),'wb') as w:w.setnchannels(2);w.setsampwidth(2);w.setframerate(sr);w.writeframes((st*32767).astype('<i2').tobytes())
 print(name,duration,'seconds')
