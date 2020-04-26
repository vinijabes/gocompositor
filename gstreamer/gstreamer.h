#ifndef GST_H
#define GST_H

#include <glib.h>
#include <gst/gst.h>
#include <gst/app/gstappsink.h>
#include <gst/app/gstappsrc.h>
#include <stdint.h>
#include <stdlib.h>

void gstreamer_init();

GstElement* gstreamer_element_factory_make(const char* factoryname,const char* name);
void gstream_element_link(GstElement* src, GstElement* dest);
void gstream_object_set(gpointer object,const char* first_property_name, ...);

GstPipeline* gstream_create_pipeline(const char* name); 

#endif