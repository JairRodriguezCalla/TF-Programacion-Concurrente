# 🔌 Predicción de Categoría Tarifaria de Suministro Eléctrico

Este proyecto implementa una solución distribuida y concurrente para la predicción automatizada de la **categoría tarifaria (`COD_TARIFA`)** de los suministros eléctricos en el Perú, utilizando datos oficiales de OSINERGMIN, un modelo de Machine Learning (Random Forest) y tecnologías como **Go, Redis y Docker**.

---

## 🎯 Objetivo del Proyecto

- Automatizar la clasificación tarifaria usando aprendizaje automático.
- Procesar grandes volúmenes de datos (>1 millón de registros) de manera eficiente.
- Implementar procesamiento concurrente y distribuido con **goroutines**, **channels** y **Redis**.
- Integrar una SPA (Next.js) con una API REST escrita en Go.
- Desplegar los servicios con **Docker Compose**.

---

## 🧠 Tecnologías Utilizadas

| Componente          | Tecnología         |
|---------------------|--------------------|
| Backend API         | Go (Golang)        |
| Modelo ML (prototipo)| Python + Scikit-learn (para reglas de referencia) |
| Concurrencia        | Go: goroutines + channels |
| Comunicación        | Redis (buffer de tareas y resultados) |
| Contenerización     | Docker / Docker Compose |
| Frontend (fase TF)  | Next.js (SPA)      |
| Persistencia (fase TF)| MongoDB          |

---
