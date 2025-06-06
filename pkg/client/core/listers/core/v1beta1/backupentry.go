// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	corev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
)

// BackupEntryLister helps list BackupEntries.
// All objects returned here must be treated as read-only.
type BackupEntryLister interface {
	// List lists all BackupEntries in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*corev1beta1.BackupEntry, err error)
	// BackupEntries returns an object that can list and get BackupEntries.
	BackupEntries(namespace string) BackupEntryNamespaceLister
	BackupEntryListerExpansion
}

// backupEntryLister implements the BackupEntryLister interface.
type backupEntryLister struct {
	listers.ResourceIndexer[*corev1beta1.BackupEntry]
}

// NewBackupEntryLister returns a new BackupEntryLister.
func NewBackupEntryLister(indexer cache.Indexer) BackupEntryLister {
	return &backupEntryLister{listers.New[*corev1beta1.BackupEntry](indexer, corev1beta1.Resource("backupentry"))}
}

// BackupEntries returns an object that can list and get BackupEntries.
func (s *backupEntryLister) BackupEntries(namespace string) BackupEntryNamespaceLister {
	return backupEntryNamespaceLister{listers.NewNamespaced[*corev1beta1.BackupEntry](s.ResourceIndexer, namespace)}
}

// BackupEntryNamespaceLister helps list and get BackupEntries.
// All objects returned here must be treated as read-only.
type BackupEntryNamespaceLister interface {
	// List lists all BackupEntries in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*corev1beta1.BackupEntry, err error)
	// Get retrieves the BackupEntry from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*corev1beta1.BackupEntry, error)
	BackupEntryNamespaceListerExpansion
}

// backupEntryNamespaceLister implements the BackupEntryNamespaceLister
// interface.
type backupEntryNamespaceLister struct {
	listers.ResourceIndexer[*corev1beta1.BackupEntry]
}
