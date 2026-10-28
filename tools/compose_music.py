"""Render an original soft looping celesta accompaniment, no external recordings."""
import numpy as np,wave,os
from pathlib import Path
sr=44100;beat=.625;n=int(sr*beat*32);mix=np.zeros((n,2),dtype=np.float64)
def note(midi,start,length,amp,pan=0,kind='bell'):
 t=np.arange(int(length*sr))/sr;f=440*2**((midi-69)/12)
 if kind=='bell':
  sound=(np.sin(2*np.pi*f*t)+.23*np.sin(2*np.pi*f*2*t)+.06*np.sin(2*np.pi*f*3*t))*np.exp(-t*3.8)
 else:sound=(np.sin(2*np.pi*f*t)+.16*np.sin(2*np.pi*f*2*t))*np.exp(-t*1.8)
 env=np.minimum(t/.012,1)*np.minimum((length-t)/.06,1);sound*=env*amp
 idx=(int(start*sr)+np.arange(len(t)))%n
 for ch,gain in enumerate([np.sqrt((1-pan)/2),np.sqrt((1+pan)/2)]):np.add.at(mix[:,ch],idx,sound*gain)
melodies=[[72,76,79,76,74,76,79,0],[69,72,76,72,71,72,76,0],[69,72,77,76,72,69,72,0],[67,71,74,71,69,71,74,0],[76,79,84,79,76,74,72,0],[76,72,69,72,76,79,76,0],[77,76,72,69,72,76,77,0],[74,71,67,71,74,71,72,0]]
chords=[[48,60,64,67],[45,57,60,64],[41,53,57,60],[43,55,59,62]]*2
for bar,mel in enumerate(melodies):
 for i,m in enumerate(mel):
  if m:note(m,(bar*4+i*.5)*beat,1.5,.095,(-.2 if i%2 else .2))
 chord=chords[bar]
 note(chord[0],bar*4*beat,2,.105,0,'bass');note(chord[0]+12,(bar*4+2)*beat,1.5,.065,0,'bass')
 for i,m in enumerate(chord[1:]):note(m,(bar*4+i*.8)*beat,2,.032,(i-1)*.25)
peak=np.max(np.abs(mix));mix*=.70/peak
p=str(Path(__file__).resolve().parents[1]/'assets/music/fox_garden.wav')
with wave.open(p,'wb') as w:w.setnchannels(2);w.setsampwidth(2);w.setframerate(sr);w.writeframes((mix*32767).astype('<i2').tobytes())
print('Music:',n/sr,'seconds; peak',np.max(np.abs(mix)),'; seam delta',np.max(np.abs(mix[0]-mix[-1])))
