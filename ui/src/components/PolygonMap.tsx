import React, { useEffect, useState } from "react"; 
import { createUseStyles } from 'react-jss';
import { FeatureGroup, MapContainer, TileLayer, Polygon, useMapEvent, Marker } from "react-leaflet";
import { EditControl } from "react-leaflet-draw";
import { LatLngLiteral } from "leaflet";

import "leaflet/dist/leaflet.css";
import "leaflet-draw/dist/leaflet.draw.css";

const useStyles = createUseStyles({
  mapContainerWrapper: {
    marginBottom: '0.5rem',
  },
  mapContainer: {
    height: '35rem',
    marginBottom: '0.5rem',
    borderRadius: '8px',
  },
});

export const polygonColors = [
  "rgba(52, 152, 219, 0.5)",
  "rgba(46, 204, 113, 0.5)",
  "rgba(241, 196, 15, 0.5)",
  "rgba(230, 126, 34, 0.5)",
];

export type PolygonMapLayer = {
  id: number;
  latlngs: LatLngLiteral[];
};

type Props = {
  layers: PolygonMapLayer[];
  setLayers: React.Dispatch<React.SetStateAction<PolygonMapLayer[]>>;
  markers: LatLngLiteral[];
}

const defaultCenter: LatLngLiteral = { lat: 51.477, lng: 0 };
const defaultZoom = 4;

export const PolygonMap: React.FC<Props> = ({ layers, setLayers, markers }) => {
  const classes = useStyles();

  const [center, setCenter] = useState<LatLngLiteral>(() => {
    const savedCenter = localStorage.getItem("polygonMapCenter");
    return savedCenter ? JSON.parse(savedCenter) : defaultCenter;
  });

  const [zoom, setZoom] = useState<number>(() => {
    const savedZoom = localStorage.getItem("polygonMapZoom");
    return savedZoom ? JSON.parse(savedZoom) : defaultZoom;
  });

  useEffect(() => {
    localStorage.setItem("polygonMapCenter", JSON.stringify(center));
  }, [center]);

  useEffect(() => {
    localStorage.setItem("polygonMapZoom", JSON.stringify(zoom));
  }, [zoom]);

  const UpdateCenterAndZoom = () => {
    const map = useMapEvent('moveend', () => {
      setCenter(map.getCenter());
      setZoom(map.getZoom());
    });
    return null;
  }

  const onCreate = (e: any) => {
    const { layerType, layer } = e;
    if (layerType === "polygon") {
      const { _leaflet_id } = layer;
      setLayers((layers) => [
        ...layers,
        { id: _leaflet_id, latlngs: layer.getLatLngs()[0] },
      ]);
    }
  };

  const onEdited = (e: any) => {
    const {
      layers: { _layers },
    } = e;
    Object.values(_layers).forEach(({ _leaflet_id, editing }: any) => {
      setLayers((layers) =>
        layers.map((l) =>
          l.id === _leaflet_id
            ? { ...l, latlngs: editing.latlngs[0] }
            : l
        )
      );
    });
  };

  const onDeleted = (e: any) => {
    const {
      layers: { _layers },
    } = e;
    Object.values(_layers).forEach(({ _leaflet_id }: any) => {
      setLayers((layers) => layers.filter((l) => l.id !== _leaflet_id));
    });
  };

  return (
    <div className={`${classes.mapContainerWrapper} roll-in`}>
      <MapContainer
        className={classes.mapContainer}
        center={center}
        zoom={zoom}
        scrollWheelZoom={false}
      >
        <UpdateCenterAndZoom />
        {markers.map((marker, index) => (
          <Marker key={index} position={marker} />
        ))}
        <FeatureGroup>
          <EditControl
            position="topleft"
            onCreated={onCreate}
            onEdited={onEdited}
            onDeleted={onDeleted}
            draw={{
              rectangle: false,
              polyline: false,
              circle: false,
              circlemarker: false,
              marker: false,
            }}
          />
            {layers.map((layer, i) => (
              <Polygon 
                key={layer.id} 
                positions={layer.latlngs} 
                pathOptions={{ color: polygonColors[(i % polygonColors.length)] }}
              />
            ))}
        </FeatureGroup>
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
      </MapContainer>
    </div>
  )
};

export const isPointInPolygon = (point: LatLngLiteral, polygon: LatLngLiteral[]): boolean => {
  let inside = false;
  const { lat, lng } = point;
  const n = polygon.length;

  for (let i = 0, j = n - 1; i < n; j = i++) {
    const xi = polygon[i].lat, yi = polygon[i].lng;
    const xj = polygon[j].lat, yj = polygon[j].lng;

    const intersect = (yi > lng) !== (yj > lng) &&
      (lat < (xj - xi) * (lng - yi) / (yj - yi) + xi);

    if (intersect) inside = !inside;
  }

  return inside;
};
