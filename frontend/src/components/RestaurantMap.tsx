import { GoogleMap, MarkerF, useJsApiLoader } from '@react-google-maps/api';

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

export function RestaurantMap({ latitude, longitude, name }: RestaurantMapProps) {
  const { isLoaded, loadError } = useJsApiLoader({
    googleMapsApiKey: import.meta.env.VITE_GOOGLE_MAPS_API_KEY || '',
    version: 'weekly',
    preventGoogleFontsLoading: true,
  });

  const center = { lat: latitude, lng: longitude };

  if (loadError) {
    return (
      <div className="w-full h-[300px] bg-(--surface-hover) rounded-lg flex items-center justify-center">
        <p className="text-(--text-muted)">Failed to load map</p>
      </div>
    );
  }

  if (!isLoaded) {
    return (
      <div className="w-full h-[300px] bg-(--surface-hover) rounded-lg flex items-center justify-center animate-pulse">
        <p className="text-(--text-muted)">Loading map...</p>
      </div>
    );
  }

  return (
    <div className="relative">
      <GoogleMap
        mapContainerStyle={containerStyle}
        center={center}
        zoom={15}
      >
        <MarkerF
          position={center}
          title={name}
        />
      </GoogleMap>
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
