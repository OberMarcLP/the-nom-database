import { useEffect, useRef, useState } from 'react';
import { GoogleMap, useJsApiLoader, type Libraries } from '@react-google-maps/api';
import { getRuntimeConfig } from '../services/api';

interface RestaurantMapProps {
  latitude: number;
  longitude: number;
  name: string;
}

const containerStyle = {
  width: '100%',
  height: '300px',
  borderRadius: '0.5rem',
};

// Advanced markers require the marker library and a map ID
const MAP_LIBRARIES: Libraries = ['marker'];

function MapPlaceholder({ text, pulse }: { text: string; pulse?: boolean }) {
  return (
    <div
      className={`w-full h-[300px] bg-(--surface-hover) rounded-lg flex items-center justify-center${pulse ? ' animate-pulse' : ''}`}
    >
      <p className="text-(--text-muted)">{text}</p>
    </div>
  );
}

function MapCanvas({
  apiKey,
  mapId,
  latitude,
  longitude,
  name,
}: RestaurantMapProps & { apiKey: string; mapId: string }) {
  const { isLoaded, loadError } = useJsApiLoader({
    googleMapsApiKey: apiKey,
    version: 'weekly',
    preventGoogleFontsLoading: true,
    libraries: MAP_LIBRARIES,
  });
  const markerRef = useRef<google.maps.marker.AdvancedMarkerElement | null>(null);

  const center = { lat: latitude, lng: longitude };

  const handleLoad = (map: google.maps.Map) => {
    markerRef.current = new google.maps.marker.AdvancedMarkerElement({
      map,
      position: center,
      title: name,
    });
  };

  const handleUnmount = () => {
    if (markerRef.current) {
      markerRef.current.map = null;
      markerRef.current = null;
    }
  };

  if (loadError) {
    return <MapPlaceholder text="Failed to load map" />;
  }

  if (!isLoaded) {
    return <MapPlaceholder text="Loading map..." pulse />;
  }

  return (
    <div className="relative">
      <GoogleMap
        mapContainerStyle={containerStyle}
        center={center}
        zoom={15}
        options={{ mapId }}
        onLoad={handleLoad}
        onUnmount={handleUnmount}
      />
      <a
        href={`https://www.google.com/maps/dir/?api=1&destination=${latitude},${longitude}`}
        target="_blank"
        rel="noopener noreferrer"
        className="absolute bottom-4 right-4 btn-glass-primary text-sm shadow-2xl"
      >
        Get Directions
      </a>
    </div>
  );
}

export function RestaurantMap({ latitude, longitude, name }: RestaurantMapProps) {
  // The key comes from the backend at runtime (published images have no
  // build-time key); the Vite env var stays as a local-dev fallback.
  const [config, setConfig] = useState<{ key: string; mapId: string } | null>(null);

  useEffect(() => {
    let cancelled = false;
    const envKey = import.meta.env.VITE_GOOGLE_MAPS_API_KEY || '';
    getRuntimeConfig().then(c => {
      if (!cancelled) {
        setConfig({
          key: c.google_maps_api_key || envKey,
          mapId: c.google_maps_map_id || 'DEMO_MAP_ID',
        });
      }
    });
    return () => {
      cancelled = true;
    };
  }, []);

  if (!config) {
    return <MapPlaceholder text="Loading map..." pulse />;
  }

  if (!config.key) {
    return <MapPlaceholder text="Map unavailable (no Google Maps API key configured)" />;
  }

  return (
    <MapCanvas apiKey={config.key} mapId={config.mapId} latitude={latitude} longitude={longitude} name={name} />
  );
}
